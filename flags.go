package main

import (
	"fmt"
	"github.com/codegangsta/cli"
)

var logLevelFlag = cli.StringFlag{
	Name:  "log, l",
	Value: "warn",
	Usage: "Log verbose output (debug, info, warn).",
}

var nameFlag = cli.StringFlag{
	Name:  "name",
	Value: "default",
	Usage: "Name of the cluster.",
}

var serversFlag = cli.StringSliceFlag{
	Name:  "servers",
	Value: &cli.StringSlice{},
	Usage: fmt.Sprintf("Initial servers. List of docker listening address: tcp://<ip1>:<port1>,tcp://<ip2>:<port2>.\n\tDefault is localhost ($DOCKER_HOST or unix:///var/run/docker.sock)."),
}

var controllerNameFlag = cli.StringFlag{
	Name:  "name",
	Usage: "Name of the project.",
}

var daemonFlag = cli.BoolFlag{
	Name:  "d",
	Usage: "Run in daemon mode.",
}
