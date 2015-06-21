package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"

	"github.com/francisbouvier/pipes/src/builder"
	"github.com/francisbouvier/pipes/src/controller"
	"github.com/francisbouvier/pipes/src/discovery"
	_ "github.com/francisbouvier/pipes/src/store/etcd"
)

var (
	Author       = "Francis Bouvier <francis.bouvier@gmail.com>"
	Contributors = []string{
		"Charles Raimbert <charles.raimbert@gmai.com>",
	}
	VERSION = "0.1.0"
)

func main() {

	app := cli.NewApp()

	app.Name = "pipes"
	app.Author = strings.Join(append([]string{Author}, Contributors...), "\n   ")
	app.Version = VERSION
	app.Usage = "A micro-services framework"
	app.Flags = []cli.Flag{logLevelFlag}

	app.Before = func(c *cli.Context) error {
		switch c.String("log") {
		case "debug":
			log.SetLevel(log.DebugLevel)
		case "info":
			log.SetLevel(log.InfoLevel)
		default:
			log.SetLevel(log.WarnLevel)
		}
		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:  "init",
			Usage: "Initiate a cluster",
			Flags: []cli.Flag{nameFlag, serversFlag},
			Action: func(c *cli.Context) {

				err := discovery.Initialize(c)
				if err != nil {
					log.Fatalln(err)
				}
			},
		},
		{
			Name:  "build",
			Usage: "Build a micro-service",
			Flags: []cli.Flag{nameFlag, serversFlag},
			Action: func(c *cli.Context) {
				if err := builder.BuildDockerImagesFromExec(c.Args(), c); err != nil {
					log.Fatalln(err)
				}
			},
		},
		{
			Name:  "run",
			Usage: "Run a workfow",
			Flags: []cli.Flag{daemonFlag, controllerNameFlag},
			Action: func(c *cli.Context) {
				if err := controller.Run(c); err != nil {
					log.Fatalln(err)
				}
			},
		},
		{
			Name:  "rm",
			Usage: "Remove workfow",
			Action: func(c *cli.Context) {
				if err := controller.Remove(c); err != nil {
					log.Fatalln(err)
				}
			},
		},
		{
			Name:  "version",
			Usage: "pipes version",
			Action: func(c *cli.Context) {
				fmt.Println("pipes", VERSION)
			},
		},
	}

	app.Run(os.Args)
}
