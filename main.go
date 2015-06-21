package main

import (
	"github.com/francisbouvier/pipes/src/builder"
	"os"
)

func main() {

	execPath_type_map := builder.AssociateExecWithType(os.Args[1:])

	builder.BuildDockerImagesFromExec(&execPath_type_map)

}
