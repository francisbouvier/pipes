package main

import (
	"github.com/francisbouvier/pipes/src/builder"
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

var (
	Author       = "Francis Bouvier <francis.bouvier@gmail.com>"
	Contributors = []string{
		"Charles Raimbert <charles.raimbert@gmai.com>",
	}
	VERSION = "0.1.0"
)

func main() {

	execPath_type_map := builder.AssociateExecWithType(os.Args[1:])

	builder.BuildDockerImagesFromExec(&execPath_type_map)

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
			Name:  "version",
			Usage: "pipes version",
			Action: func(c *cli.Context) {
				fmt.Println("pipes", VERSION)
			},
		},
	}

	app.Run(os.Args)
}
