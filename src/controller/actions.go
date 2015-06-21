package controller

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/francisbouvier/pipes/src/discovery"
	"github.com/francisbouvier/pipes/src/orch/swarm"
)

func Run(c *cli.Context) error {
	daemon := c.Bool("d")
	if daemon {
		log.Debugln("Running project in daemon")
	} else {
		log.Debugln("Running project")
	}

	// Services
	if len(c.Args()) == 0 {
		msg := fmt.Sprintf(
			"You need to provide a workflow, ie. %s",
			"\"service1 | service2 | service3\"",
		)
		return errors.New(msg)
	}
	services := []string{}
	for _, service := range strings.Split(c.Args()[0], "|") {
		service = strings.TrimSpace(service)
		services = append(services, service)
	}
	// TODO: check if services exists in store
	log.Debugln("Services", services)

	// Project
	name := c.String("name")
	st, err := discovery.GetStore(c)
	if err != nil {
		return err
	}
	p, err := NewProject(name, st)
	if err != nil {
		return err
	}
	log.Debugln(p)
	if err = p.SetServices(services); err != nil {
		return err
	}
	log.Debugf("Project %s (%s)\n", p.ID, p.Name)
	if daemon {
		fmt.Printf("Project %s (%s)\n", p.ID, p.Name)
	}

	// Run
	o, err := swarm.New(st)
	if err != nil {
		return err
	}
	ctr := Controller{orch: o, project: p}
	api, err := ctr.LaunchAPI()
	if err != nil {
		return err
	}
	for _, service := range ctr.project.Services {
		if err = ctr.launchService(service); err != nil {
			return err
		}
	}
	log.Debugf("API listening on: http://%s\n", api.Addr())

	if daemon {
		fmt.Printf("API listening on: http://%s\n", api.Addr())
		return nil
	}

	// Query
	query := c.Args()[1:]
	if len(query) > 0 {
		if err = ctr.project.Query(strings.Join(query, " ")); err != nil {
			return err
		}
	}

	// Stop
	if err = ctr.Stop(); err != nil {
		return nil
	}

	return nil
}

func Remove(c *cli.Context) error {

	// Project
	st, err := discovery.GetStore(c)
	if err != nil {
		return err
	}
	var id string
	if len(c.Args()) == 0 {
		// Get main project
		id, err = st.Read("main_project", "")
		if err != nil {
			return err
		}
	} else {
		id = c.Args()[0]
		// Try by name
		if realId, err := st.Read(id, "names"); err == nil {
			id = realId
		}
	}
	p, err := GetProject(id, st)
	if err != nil {
		return errors.New("Project does not exists")
	}
	if !p.Running() {
		return errors.New("Project is not running")
	}

	// Controller
	o, err := swarm.New(st)
	if err != nil {
		return err
	}
	ctr := Controller{orch: o, project: p}

	// Stop
	if err = ctr.Stop(); err != nil {
		return nil
	}
	fmt.Println("Project deleted:", p.ID)

	return nil
}
