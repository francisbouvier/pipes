package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/francisbouvier/pipes/src/discovery"
	_ "github.com/francisbouvier/pipes/src/store/etcd"
)

func main() {

	app := cli.NewApp()

	app.Name = "pipes"
	app.Author = "Francis Bouvier <francis.bouvier@gmail.com>"
	app.Version = "0.1.0"
	app.Usage = "A micro-services framework"
	app.Flags = []cli.Flag{logLevelFlag, serviceFlag}

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
				if err := discovery.Initialize(c); err != nil {
					log.Fatalln(err)
				}
			},
		},
	}

	app.Run(os.Args)
}
