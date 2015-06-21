package main

import (
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/francisbouvier/pipes/src/controller"
	"github.com/francisbouvier/pipes/src/store/etcd"
	"github.com/francisbouvier/pipes/src/wrapper"
	"github.com/julienschmidt/httprouter"
)

func launch(storeAddr, projectID, addr string) error {

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

	// Private client
	private, err := wrapper.ConnectWamp(routerAddr, "realm1")
	if err != nil {
		return err
	}

	// Wrapper
	w := wrapper.New("api", project, private)
	if err = w.Init(); err != nil {
		return err
	}

	// Handler and router
	h := NewHandler(w)
	router := httprouter.New()
	router.POST("/", h.httpAPI)
	router.GET("/jobs/:id/", h.httpJob)

	// Server
	log.Infoln("Serving API on:", addr)
	return http.ListenAndServe(addr, router)
}

var logLevelFlag = cli.StringFlag{
	Name:  "log, l",
	Value: "info",
	Usage: "Log verbose output (debug, info, warn).",
}

func main() {

	app := cli.NewApp()

	app.Name = "pipes_api"
	app.Author = "Francis Bouvier <francis.bouvier@gmail.com>"
	app.Version = "0.1.0"
	app.Usage = "API for pipes, micro-services framework"
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
		addr := "0.0.0.0:8080"
		if err := launch(storeAddr, projectID, addr); err != nil {
			log.Fatalln(err)
		}
	}
	app.Run(os.Args)
}
