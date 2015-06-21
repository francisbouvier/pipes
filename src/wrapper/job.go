package wrapper

import (
	"math/rand"

	log "github.com/Sirupsen/logrus"
	"github.com/francisbouvier/wampace/wamp"
)

type service struct {
	name string
	uri  string
	wait bool
}

type status struct {
	Code    int
	Message string
}

type Job struct {
	ID        int
	services  map[string]service
	rc        map[string]*wamp.RCall
	Responses []interface{}
	Finish    chan bool
	status    status
}

func (j *Job) Finished(code int) {
	switch code {
	case 2:
		j.status = status{Code: 2, Message: "Success"}
	case 3:
		j.status = status{Code: 3, Message: "Error"}
	}
}

func (j *Job) Status() status {
	return j.status
}

func (j *Job) result(args []interface{}, kwargs map[string]interface{}) {
	log.Debugln("Result call with args:", args)
	// Consolidize reponses
	// Starting from here we are going outside the Wamp protocole definition
	// with len(args) fixe to 0
	// and kwargs empty
	j.Responses = append(j.Responses, args[0])
	if len(j.Responses) == len(j.services) {
		code := 2
		for _, resp := range j.Responses {
			switch resp.(type) {
			case map[string]interface{}:
				for k, _ := range resp.(map[string]interface{}) {
					if k == "error" {
						code = 3
						break
					}
				}
			}
			if code == 3 {
				break
			}
		}
		j.Finish <- true
		j.Finished(code)
	}
}

func (j *Job) call(c *wamp.Client, args []interface{}) {
	for _, s := range j.services {
		go func() {
			rc := c.Call(s.uri, args, map[string]interface{}{})
			<-rc.Result
			j.result(rc.Args, rc.Kwargs)
		}()
	}
	j.status = status{Code: 1, Message: "Started"}
}

func NewJob(services map[string]service) (j *Job) {
	id := rand.Int()
	log.Debugln("Job ID:", id)
	finish := make(chan bool, 1)
	j = &Job{
		ID:       id,
		services: make(map[string]service),
		Finish:   finish,
		status:   status{Code: 0, Message: "Not started"},
	}
	for k, v := range services {
		j.services[k] = v
		// On ajoute v.args et v.kwargs issus des args de la CLI
	}
	log.Debugln("Job services:", j.services)
	return
}
