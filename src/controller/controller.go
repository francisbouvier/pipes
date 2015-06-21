package controller

import (
	"github.com/francisbouvier/pipes/src/orch"
)

type Controller struct {
	orch    orch.Orch
	project *project
}
