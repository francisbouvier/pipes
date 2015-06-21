package wrapper

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	// "strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/francisbouvier/pipes/src/controller"
	"github.com/francisbouvier/pipes/src/store"
	"github.com/francisbouvier/wampace/client"
	"github.com/francisbouvier/wampace/wamp"
)

type Wrapper struct {
	name     string
	project  *controller.Project
	st       store.Store
	services map[string]service
	jobs     map[int]*Job
	c        *wamp.Client
	Cmd      string
	Mode     string
}

func (w *Wrapper) Init() error {
	dir := fmt.Sprintf("projects/%s/services/%s", w.project.ID, w.name)
	list, err := w.st.List("next", dir)
	if err != nil {
		return errors.New("No services following: " + w.name)
	}
	dir = fmt.Sprintf("projects/%s/services/%s/next", w.project.ID, w.name)
	for _, serviceName := range list {
		s := service{name: serviceName}
		s.uri = fmt.Sprintf("com.%s.%s", w.project.ID, serviceName)
		w.services[serviceName] = s
	}
	log.Debugln("Wrapper services:", w.services)
	return nil
}

func (w *Wrapper) Handle(args []interface{}) (job *Job) {
	job = NewJob(w.services)
	w.jobs[job.ID] = job
	job.call(w.c, args)
	log.Infoln("Launch job:", job.ID)
	return job
}

func argsBin(cmd string, cmdArgs []string, args []interface{}) ([]interface{}, error) {
	resp := []interface{}{}
	for _, elem := range args {
		arg := elem.(string)
		argList := strings.Split(arg, " ")
		cmdArgs = append(cmdArgs, argList...)
	}
	bin := exec.Command(cmd, cmdArgs...)
	res, err := bin.Output()
	if err != nil {
		return resp, err
	}
	bin.Wait()
	arg := string(res)
	arg = strings.TrimSuffix(arg, "\n")
	log.Debugln("Result:", arg)
	resp = append(resp, arg)
	return resp, nil
}

func stdinBin(cmd string, cmdArgs []string, args []interface{}) ([]interface{}, error) {
	resp := []interface{}{}
	log.Debugln("cmd:", cmd)
	log.Debugln("cmdArgs:", cmdArgs)
	bin := exec.Command(cmd, cmdArgs...)
	in, err := bin.StdinPipe()
	if err != nil {
		return resp, err
	}
	out, err := bin.StdoutPipe()
	if err != nil {
		log.Debugln("Error", err)
		return resp, err
	}
	bin.Start()
	inArg := args[0].(string)
	log.Debugln("inArg:", inArg)
	in.Write([]byte(inArg + "\n"))
	in.Close()
	res, err := ioutil.ReadAll(out)
	if err != nil {
		log.Debugln("Error", err)
		return resp, err
	}
	bin.Wait()
	arg := string(res)
	arg = strings.TrimSuffix(arg, "\n")
	log.Debugln("Resp:", arg)
	resp = append(resp, arg)
	return resp, nil
}

func (w *Wrapper) Procedure(args []interface{}, kwargs map[string]interface{}) (resp []interface{}, k map[string]interface{}) {
	log.Infoln("Receive call with args:", args)

	// Launch binary
	resp = []interface{}{}
	k = map[string]interface{}{}
	fullCmd := strings.Split(w.Cmd, " ")
	cmd := fullCmd[0]
	cmdArgs := fullCmd[1:]
	var err error
	log.Debugln("Mode:", w.Mode)
	if w.Mode == "args" {
		resp, err = argsBin(cmd, cmdArgs, args)
	} else if w.Mode == "stdin" {
		resp, err = stdinBin(cmd, cmdArgs, args)
	}
	if err != nil {
		e := map[string]interface{}{"error": err.Error()}
		resp = []interface{}{e}
		return
	}

	// Launch next jobs
	job := w.Handle(resp)
	if len(job.services) > 0 {
		<-job.Finish
		resp = job.Responses
	}
	job.Finished(2)
	return
}

func New(name string, project *controller.Project, c *wamp.Client) (w *Wrapper) {
	w = &Wrapper{
		name:     name,
		project:  project,
		st:       project.Store,
		services: map[string]service{},
		jobs:     map[int]*Job{},
		c:        c,
	}
	return
}

func ConnectWamp(url, realm string) (c *wamp.Client, err error) {
	c, err = client.New(url)
	if err != nil {
		return
	}
	err = c.Join(realm)
	return
}
