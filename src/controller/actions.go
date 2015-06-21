package controller

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/francisbouvier/pipes/src/discovery"
	"github.com/francisbouvier/pipes/src/orch/swarm"
	"github.com/francisbouvier/pipes/src/store"
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
	var query string
	for i, service := range strings.Split(c.Args()[0], "|") {
		if i == 0 {
			serviceFull := strings.Split(service, " ")
			service = serviceFull[0]
			if len(serviceFull) > 1 {
				query = strings.Join(serviceFull[1:], " ")
			}
		}
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

	// Controller
	o, err := swarm.New(st)
	if err != nil {
		return err
	}
	ctr := Controller{orch: o, project: p}

	// Run
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
	if query != "" {
		if err = ctr.project.Query(query); err != nil {
			return err
		}
	}

	// Stop
	if err = ctr.Stop(); err != nil {
		return nil
	}

	return nil
}

func getProject(args []string, st store.Store) (*Project, error) {
	var id string
	var err error
	if len(args) == 0 {
		// Get main project
		id, err = st.Read("main_project", "")
		if err != nil {
			return nil, err
		}
	} else {
		id = args[0]
		// Try by name
		if realId, err := st.Read(id, "names"); err == nil {
			id = realId
		}
	}
	p, err := GetProject(id, st)
	if err != nil {
		return nil, errors.New("Project does not exists")
	}
	if !p.Running() {
		return nil, errors.New("Project is not running")
	}
	log.Debugln("Project:", p.ID)
	return p, nil
}

func Query(c *cli.Context) error {

	// Project
	st, err := discovery.GetStore(c)
	if err != nil {
		return err
	}
	args := []string{}
	if c.String("name") != "" {
		args = append(args, c.String("name"))
	}
	p, err := getProject(args, st)
	if err != nil {
		return err
	}

	// Controller
	o, err := swarm.New(st)
	if err != nil {
		return err
	}
	ctr := Controller{orch: o, project: p}

	// Query
	query := strings.Join(c.Args(), " ")
	log.Infoln("Query for:", query)
	if err = ctr.project.Query(query); err != nil {
		return err
	}

	return nil
}

func Remove(c *cli.Context) error {

	// Project
	st, err := discovery.GetStore(c)
	if err != nil {
		return err
	}
	p, err := getProject(c.Args(), st)
	if err != nil {
		return nil
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
