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

var serviceFlag = cli.StringFlag{
	Name:  "service, s",
	Value: "local",
	Usage: "Service discovery.",
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
