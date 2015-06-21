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
	log.Debugln("Running project")

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
	fmt.Printf("Project %s (%s)\n", p.ID, p.Name)

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
	fmt.Printf("API listening on: http://%s\n", api.Addr())

	return nil
}
