package engine

import (
	"fmt"
)

type Image struct {
	Id     string
	Name   string
	Manual bool
}

type Container struct {
	Id          string
	Name        string
	Hostname    string
	Image       Image
	IP          string
	Ports       []map[string]string
	Cmd         []string
	Tty         bool
	Env         []string
	Active      bool
	NetworkMode string
	Gateway     string
}

// Etat démarré, pourrait être chargé depuis un json
type Pod struct {
	Image      Image
	Containers []Container
	// Contraints []contraint
	// RunRules map[*container]map[string]string
}

func (cont Container) Addr() (addr string) {
	if len(cont.Ports) > 0 {
		for _, hostPort := range cont.Ports[0] {
			addr = fmt.Sprintf("%s:%s", cont.IP, hostPort)
			break
		}
	} else {
		addr = cont.IP
	}
	return
}

type Engine interface {
	Run(*Container) error
	Stop(*Container) error
	Remove(*Container) error
	List() ([]*Container, error)
	GetImg(string) (Image, error)
	PullImg(string) (Image, error)
	BuildImg(string, string) (Image, error)
	RemoveImg(string) error
}
