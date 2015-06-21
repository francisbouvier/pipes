package swarm

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/francisbouvier/pipes/src/engine"
	"github.com/francisbouvier/pipes/src/engine/docker"
	"github.com/francisbouvier/pipes/src/store"
	"github.com/francisbouvier/pipes/src/utils"
)

const IMAGE = "swarm:0.3.0"

type Swarm struct {
	Store  store.Store
	engine docker.Docker
}

func (sw Swarm) Run(cont *engine.Container) (err error) {
	// Add image affinity to ensure that image is on same node
	aff := fmt.Sprintf("affinity:image==%s", cont.Image.Name)
	cont.Env = append(cont.Env, aff)
	return sw.engine.Run(cont)
}

func (sw Swarm) Stop(cont *engine.Container) error {
	return sw.engine.Stop(cont)
}

func (sw Swarm) Remove(cont *engine.Container) error {
	return sw.engine.Remove(cont)
}

func (sw Swarm) List() (conts []*engine.Container, err error) {
	return sw.engine.List()
}

func (sw Swarm) GetImg(name string) (img engine.Image, err error) {
	return sw.engine.GetImg(name)
}

func (sw Swarm) PullImg(name string) (img engine.Image, err error) {
	return sw.engine.PullImg(name)
}

func (sw Swarm) BuildImg(name, dir string) (img engine.Image, err error) {
	return sw.engine.BuildImg(name, dir)
}

func (sw Swarm) RemoveImg(name string) (err error) {
	// TODO: Right now Swarm doesn't handle removing images
	// return sw.engine.RemoveImg(name)
	return nil
}

func (sw Swarm) manager(server string, eng docker.Docker) (*engine.Container, error) {
	log.Debugf("Installing Swarm manager on node %s...\n", server)

	img := engine.Image{Name: IMAGE}
	name := "swarm_manager"
	port := "2375"
	ip := utils.AddrToIP(server)
	cmd := []string{
		"-l", "debug",
		"manage",
		"-H", fmt.Sprintf("tcp://%s:%s", "0.0.0.0", port),
		"--addr", fmt.Sprintf("%s:%s", ip, port),
	}
	cmd = append(cmd, fmt.Sprintf("etcd://%s/cluster", utils.SplitAddr(sw.Store.Addr())))
	container := &engine.Container{
		Name:     name,
		Hostname: name,
		Image:    img,
		IP:       ip,
		Cmd:      cmd,
		Ports: []map[string]string{
			map[string]string{port: ""},
		},
	}
	err := eng.Run(container)
	fmt.Println("Swarm manager running on node:", server)
	return container, err
}

func (sw Swarm) agent(server string, eng docker.Docker) (*engine.Container, error) {
	log.Debugf("Installing Swarm agent on node %s...\n", server)

	// Image
	img := engine.Image{Name: IMAGE}

	// Run
	name := "swarm_agent"
	port := "2375"
	ip := utils.AddrToIP(server)
	cmd := []string{
		"join",
		"--addr", fmt.Sprintf("%s:%s", ip, port),
		fmt.Sprintf("etcd://%s/cluster", utils.SplitAddr(sw.Store.Addr())),
	}
	container := &engine.Container{
		Name:     name,
		Hostname: name,
		Image:    img,
		IP:       ip,
		Cmd:      cmd,
	}
	if _, err := eng.GetImg(img.Name); err != nil {
		if _, err = eng.PullImg(img.Name); err != nil {
			return nil, err
		}
	}
	err := eng.Run(container)
	fmt.Println("Swarm agent running on node:", server)
	return container, err
}

func (sw Swarm) Join(server string) error {
	eng, err := docker.New(server, "")
	if err != nil {
		return err
	}
	_, err = sw.agent(server, eng)
	return err
}

func (sw Swarm) Leave(server string) error {
	eng, err := docker.New(server, "")
	if err != nil {
		return err
	}
	containers, err := eng.List()
	if err != nil {
		return err
	}
	var swarmAgent *engine.Container
	for _, container := range containers {
		if container.Name == "swarm_agent" {
			swarmAgent = container
			break
		}
	}
	if swarmAgent != nil {
		if err = eng.Stop(swarmAgent); err != nil {
			return err
		}
		if err = eng.Remove(swarmAgent); err != nil {
			return err
		}
	}
	return nil
}

func (sw Swarm) Initialize(servers []string) (err error) {

	var clusterContainers []*engine.Container
	var engines []docker.Docker

	// Ensure the cluster is in Store
	if err = sw.Store.Write("/cluster", "", ""); err != nil {
		return err
	}

	// Nodes
	for _, server := range servers {

		// Connect to Docker
		log.Infoln("Connecting to Docker on:", server)
		eng, err := docker.New(server, "")
		engines = append(engines, eng)
		if err != nil {
			return err
		}

		// Agent
		if _, err = sw.agent(server, eng); err != nil {
			return err
		}
	}

	// Manager
	var swarmManager *engine.Container
	for _, container := range clusterContainers {
		if container.Name == "swarm_manager" {
			swarmManager = container
			break
		}
	}
	if swarmManager == nil {
		i := utils.PickServer(servers)
		if swarmManager, err = sw.manager(servers[i], engines[i]); err != nil {
			return
		}
	}

	// Connect manager
	log.Infoln("Connecting to Swarm manager on:", "tcp://"+swarmManager.Addr())
	for i := 0; i < 3; i++ {
		// We have to wait swarm manager to init
		time.Sleep(200 * time.Millisecond)
		_, err = docker.New("tcp://"+swarmManager.Addr(), "")
		if err == nil {
			break
		}
	}
	if err != nil {
		return
	}
	return sw.Store.Write("manager", swarmManager.Addr(), "cluster/docker/swarm")
}

func New(st store.Store) (sw Swarm, err error) {
	addr, err := st.Read("manager", "cluster/docker/swarm")
	if err != nil {
		return
	}
	eng, err := docker.New("tcp://"+addr, "")
	if err != nil {
		return
	}
	sw = Swarm{Store: st, engine: eng}
	return
}
