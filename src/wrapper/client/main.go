package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/francisbouvier/pipes/src/controller"
	"github.com/francisbouvier/pipes/src/store/etcd"
	"github.com/francisbouvier/pipes/src/wrapper"
)

func launch(storeAddr, projectID, service string) error {

	// Store, project and router
	st := &etcd.Etcd{}
	st.New(storeAddr)
	project, err := controller.GetProject(projectID, st)
	if err != nil {
		return err
	}
	routerAddr, err := project.Store.Read("addr", "router")
	if err != nil {
		return err
	}

	// Client
	client, err := wrapper.ConnectWamp(routerAddr, "realm1")
	if err != nil {
		return err
	}

	// Wrapper
	w := wrapper.New(service, project, client)
	dir := fmt.Sprintf("services/%s", service)
	w.Cmd, err = project.Store.Read("command", dir)
	if err != nil {
		return err
	}
	log.Debugln("Cmd:", w.Cmd)
	w.Mode, err = project.Store.Read("input_mode", dir)
	if err != nil {
		return err
	}
	if err = w.Init(); err != nil {
		return err
	}

	uri := fmt.Sprintf("com.%s.%s", project.ID, service)
	rp := client.Register(uri, w.Procedure)
	<-rp.Registred

	client.End()
	return nil
}

var logLevelFlag = cli.StringFlag{
	Name:  "log, l",
	Value: "info",
	Usage: "Log verbose output (debug, info, warn).",
}

func main() {

	app := cli.NewApp()

	app.Name = "pipes_client"
	app.Author = "Francis Bouvier <francis.bouvier@gmail.com>"
	app.Version = "0.1.0"
	app.Usage = "Client for pipes, micro-services framework"
	app.Flags = []cli.Flag{logLevelFlag}

	app.Action = func(c *cli.Context) {
		switch c.String("log") {
		case "debug":
			log.SetLevel(log.DebugLevel)
		case "warn":
			log.SetLevel(log.WarnLevel)
		default:
			log.SetLevel(log.InfoLevel)
		}

		storeAddr := c.Args()[0]
		projectID := c.Args()[1]
		service := c.Args()[2]
		if err := launch(storeAddr, projectID, service); err != nil {
			log.Fatalln(err)
		}
	}
	app.Run(os.Args)
}
