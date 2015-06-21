package controller

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/francisbouvier/pipes/src/discovery"
	"github.com/francisbouvier/pipes/src/engine"
	"github.com/francisbouvier/pipes/src/orch"
)

type Controller struct {
	orch    orch.Orch
	project *Project
}

func (ctr *Controller) LaunchAPI() (*engine.Container, error) {
	containerName := fmt.Sprintf("%s_api", ctr.project.Name)
	img := engine.Image{Name: discovery.API_IMAGE}
	port := "8080"
	container := &engine.Container{
		Name:     containerName,
		Hostname: containerName,
		Image:    img,
		Ports: []map[string]string{
			map[string]string{port: ""},
		},
		Cmd: []string{
			"-l", "debug",
			ctr.project.Store.Addr(),
			ctr.project.ID,
		},
	}
	err := ctr.orch.Run(container)
	if err != nil {
		return container, nil
	}
	dir := fmt.Sprintf("projects/%s/services/api", ctr.project.ID)
	err = ctr.project.Store.Write("addr", container.Addr(), dir)
	if err = ctr.project.SetContainer("api", container); err != nil {
		return container, err
	}
	log.Infoln("Running API on:", container.Addr())
	return container, err
}

func (ctr *Controller) launchService(service string) error {
	log.Infoln("Running:", service)
	name := fmt.Sprintf("%s_%s", ctr.project.Name, service)
	imgName := strings.Split(service, ".")[0]
	img := engine.Image{Name: imgName}

	// Run
	cmd := []string{
		"-l", "debug",
		ctr.project.Store.Addr(),
		ctr.project.ID, service,
	}
	container := &engine.Container{
		Name:     name,
		Hostname: name,
		Image:    img,
		Ports:    []map[string]string{},
		Cmd:      cmd,
	}
	if err := ctr.orch.Run(container); err != nil {
		return err
	}
	if err := ctr.project.SetContainer(service, container); err != nil {
		return err
	}
	return nil
}
