package controller

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/docker/docker/pkg/stringid"
	"github.com/francisbouvier/pipes/src/engine"
	"github.com/francisbouvier/pipes/src/store"
)

type Project struct {
	ID       string
	Name     string
	Store    store.Store
	Services []string
}

func NewProject(name string, st store.Store) (*Project, error) {
	var err error
	p := &Project{Store: st}

	// Name & ID
	if name != "" {
		if _, err = p.Store.Read(name, "names"); err == nil {
			return p, errors.New("Project already exists")
		}
	} else {
		p.Name = namesgenerator.GetRandomName(5)
		p.ID = stringid.GenerateRandomID()
	}

	// Create project in store
	dir := fmt.Sprintf("projects/%s", p.ID)
	if err = p.Store.Write("name", p.Name, dir); err != nil {
		return p, err
	}
	if err = p.Store.Write("running", "true", dir); err != nil {
		return p, err
	}
	if err = p.Store.Write("services", "", dir); err != nil {
		return p, err
	}
	if err = p.Store.Write(p.Name, p.ID, "names"); err != nil {
		return p, err
	}
	if err = p.Store.Write("main_project", p.ID, ""); err != nil {
		return p, err
	}
	return p, nil
}

func GetProject(id string, st store.Store) (*Project, error) {
	p := &Project{ID: id, Store: st}
	dir := fmt.Sprintf("projects/%s", p.ID)
	var err error
	p.Name, err = p.Store.Read("name", dir)
	if err != nil {
		return p, err
	}
	p.Services, err = p.Store.List("services", dir)
	if err != nil {
		return p, err
	}
	return p, nil
}

func (p *Project) nextService(service string, next []string, err error) ([]string, error) {
	dir := fmt.Sprintf("projects/%s/services/%s", p.ID, service)
	services, err := p.Store.List("next", dir)
	if err != nil {
		return next, err
	}
	if len(services) == 0 {
		return next, nil
	}
	next = append(next, services[0])
	return p.nextService(services[0], next, err)
}

func (p *Project) GetPipes() ([]string, error) {
	pipe := []string{}
	var err error
	pipe, err = p.nextService("api", []string{}, nil)
	if err != nil {
		return pipe, err
	}
	return pipe, nil
}

func (p *Project) SetContainer(service string, container *engine.Container) error {
	dir := fmt.Sprintf("projects/%s/services/%s/containers/", p.ID, service)
	if err := p.Store.Write(container.Id, container.IP, dir); err != nil {
		return err
	}
	return nil
}

func (p *Project) GetContainer(service string) (*engine.Container, error) {
	dir := fmt.Sprintf("projects/%s/services/%s", p.ID, service)
	contID, err := p.Store.List("containers", dir)
	if err != nil {
		return nil, err
	}
	if len(contID) == 0 {
		return nil, errors.New(fmt.Sprintf("No containers for %s\n", service))
	}
	return &engine.Container{Id: contID[0]}, nil
}

func (p *Project) RemoveContainer(service string, cont *engine.Container) error {
	dir := fmt.Sprintf("projects/%s/services/%s/containers/", p.ID, service)
	return p.Store.Delete(cont.Id, dir)
}

func (p *Project) SetServices(services []string) error {
	p.Services = services

	services = append([]string{"api"}, services...)
	for i, service := range services {
		dir := fmt.Sprintf("projects/%s/services", p.ID)
		if err := p.Store.Write(service, "", dir); err != nil {
			return err
		}
		dir = fmt.Sprintf("%s/%s", dir, service)
		if err := p.Store.Write("next", "", dir); err != nil {
			return err
		}
		// Write next service,
		// Except for last service
		if i < (len(services) - 1) {
			dir = fmt.Sprintf("%s/%s", dir, "next")
			if err := p.Store.Write(services[i+1], "", dir); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Project) Running() bool {
	dir := fmt.Sprintf("projects/%s", p.ID)
	value, err := p.Store.Read("running", dir)
	if err != nil {
		return false
	}
	status, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return status
}

func (p *Project) Query(query string) error {

	// Check running
	if running := p.Running(); running == false {
		return errors.New("Project is not running")
	}

	// API
	dir := fmt.Sprintf("projects/%s/services/api", p.ID)
	api, err := p.Store.Read("addr", dir)
	if err != nil {
		return err
	}
	// TODO: use wamp client instead of long polling

	// Post query
	form := url.Values{}
	form.Set("query", query)
	resp1, err := http.PostForm("http://"+api, form)
	if err != nil {
		return err
	}
	defer resp1.Body.Close()
	body, err := ioutil.ReadAll(resp1.Body)
	if err != nil {
		return err
	}
	job := strings.TrimPrefix(string(body), "Job ID: ")
	job = strings.TrimSuffix(job, "\n")
	log.Infof("API response: [%d] - Job %s", resp1.StatusCode, job)
	if resp1.StatusCode != 200 {
		msg := fmt.Sprintf("API error: %s", resp1.StatusCode)
		return errors.New(msg)
	}

	// Get job
	final := false
	const interval = 200
	timeout := 10
	log.Debugln("Timeout:", timeout, "seconds")
	fmt.Println("Waiting results ...")
	tries := (timeout * 1000) / interval
	for i := 0; i < tries; i++ {
		time.Sleep(interval * time.Millisecond)
		resp, err := http.Get(fmt.Sprintf("http://%s/jobs/%s/", api, job))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		content := strings.Split(strings.TrimSuffix(string(body), "\n"), "\n")
		status := strings.TrimPrefix(content[1], "Job status: ")
		log.Debugf("API response: [%d] - Result %s", resp.StatusCode, status)
		if status == "Success" {
			result := strings.TrimPrefix(content[2], "Job result: ")
			fmt.Println(result)
			if len(content) > 3 {
				for _, cont := range content[2:] {
					fmt.Println(cont)
				}
			}
			final = true
			break
		}
	}
	if final == false {
		fmt.Println("Timeout")
	}
	return nil
}

func (p *Project) Stop() error {
	dir := fmt.Sprintf("projects/%s", p.ID)
	if err := p.Store.Write("running", "false", dir); err != nil {
		return err
	}
	return nil
}
