package discovery

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/francisbouvier/pipes/src/engine"
	"github.com/francisbouvier/pipes/src/orch/swarm"
)

const IMAGE = "francisbouvier/wampace"
const API_IMAGE = "francisbouvier/pipes_api"

func wampRouter(eng swarm.Swarm) (*engine.Container, error) {
	log.Debugf("Installing Wamp Router...\n")

	// Image
	image, err := eng.GetImg(IMAGE)
	if err != nil {
		log.Debugf("Pulling image %s...\n", IMAGE)
		image, err = eng.PullImg(IMAGE)
		if err != nil {
			return nil, err
		}
	}
	log.Debugf("Image %s available...\n", IMAGE)

	// Run
	name := "wamp_router"
	container := &engine.Container{
		Name:     name,
		Hostname: name,
		Image:    image,
		Ports: []map[string]string{
			map[string]string{"1234": ""},
		},
	}
	err = eng.Run(container)
	fmt.Printf("WAMP router on node: %s\n", container.Addr())

	// Write in the etcd store
	err = eng.Store.Write("addr", container.Addr(), "router")
	if err != nil {
		return nil, err
	}

	// API Image
	_, err = eng.GetImg(API_IMAGE)
	if err != nil {
		log.Debugf("Pulling image %s...\n", API_IMAGE)
		_, err = eng.PullImg(API_IMAGE)
		if err != nil {
			return nil, err
		}
	}

	return container, err
}
