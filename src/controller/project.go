package controller

import (
	"errors"
	"fmt"

	_ "github.com/Sirupsen/logrus"
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
		return nil, err
	}
	return &engine.Container{Id: contID[0]}, nil
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
