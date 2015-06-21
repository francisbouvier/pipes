package main

import (
	"fmt"
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/francisbouvier/pipes/src/wrapper"
	"github.com/julienschmidt/httprouter"
)

type handler struct {
	wrapper *wrapper.Wrapper
	jobs    map[int]*wrapper.Job
}

func (h *handler) httpAPI(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	log.Debugln("Query:", r.Form)
	msg := []interface{}{}
	query := r.Form["query"]
	if query != nil {
		for _, elem := range query {
			msg = append(msg, elem)
		}
	}
	job := h.wrapper.Handle(msg)
	h.jobs[job.ID] = job
	fmt.Fprintf(w, "Job ID: %d\n", job.ID)
}

func (h *handler) httpJob(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	log.Debugln("Params:", params)
	id := params.ByName("id")
	ID, _ := strconv.Atoi(id)
	job, prs := h.jobs[ID]
	if prs == false {
		http.NotFound(w, r)
		return
	}
	log.Infoln("Job:", job)
	s := job.Status()
	t := fmt.Sprintf("Job ID: %d\nJob status: %s\n", job.ID, s.Message)
	switch s.Code {
	case 2:
		t = fmt.Sprintf("%sJob result: %s\n", t, job.Responses[0])
	case 3:
		r := job.Responses[0].(map[string]interface{})
		t = fmt.Sprintf("%sJob error: %s\n", t, r["error"])
	}
	fmt.Fprintf(w, t)
}

func NewHandler(w *wrapper.Wrapper) (h *handler) {
	h = &handler{
		wrapper: w,
		jobs:    map[int]*wrapper.Job{},
	}
	return
}
