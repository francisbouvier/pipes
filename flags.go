package main

import (
	"github.com/codegangsta/cli"
)

var logLevelFlag = cli.StringFlag{
	Name:  "log, l",
	Value: "warn",
	Usage: "Log verbose output (debug, info, warn).",
}
