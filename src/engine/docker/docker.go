package docker

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/francisbouvier/pipes/src/engine"
	dockerclient "github.com/fsouza/go-dockerclient"
)

type Docker struct {
	client *dockerclient.Client
}

func (d Docker) Run(cont *engine.Container) (err error) {
	contPorts := map[dockerclient.Port]struct{}{}
	hostPorts := map[dockerclient.Port][]dockerclient.PortBinding{}
	for _, m := range cont.Ports {
		for k, v := range m {
			var port dockerclient.Port
			port = dockerclient.Port(fmt.Sprintf("%s/tcp", k))
			contPorts[port] = struct{}{}
			hostPorts[port] = []dockerclient.PortBinding{
				dockerclient.PortBinding{HostIP: "", HostPort: v},
			}
		}
	}
	opts := dockerclient.CreateContainerOptions{
		Name: cont.Name,
		Config: &dockerclient.Config{
			Image:        cont.Image.Name,
			Hostname:     cont.Hostname,
			Cmd:          cont.Cmd,
			Tty:          cont.Tty,
			ExposedPorts: contPorts,
			Env:          cont.Env,
		},
	}
	hostConfig := &dockerclient.HostConfig{
		NetworkMode:  cont.NetworkMode,
		PortBindings: hostPorts,
	}
	log.Debugln("Create container:", cont.Name)
	c, err := d.client.CreateContainer(opts)
	if err != nil {
		return
	}
	err = d.client.StartContainer(c.ID, hostConfig)
	if err != nil {
		return
	}
	c, err = d.client.InspectContainer(c.ID)
	if err != nil {
		return
	}
	cont.Id = c.ID

	// Saving network informations
	if c.NetworkSettings == nil {
		return errors.New("Failed to get container network settings")
	}
	cont.Gateway = c.NetworkSettings.Gateway
	for port, bindings := range c.NetworkSettings.Ports {
		contPort := strings.TrimSuffix(string(port), "/tcp")
		for _, elem := range cont.Ports {
			hostPort, prs := elem[contPort]
			if prs {
				if hostPort != bindings[0].HostPort {
					elem[contPort] = bindings[0].HostPort
				}
				break
			}
		}
	}
	// Swarm extras
	if c.Node != nil {
		cont.IP = c.Node.IP
	}
	log.Debugln("Container:", cont)

	log.Infoln("Run:", cont.Name)
	return
}

func (d Docker) Stop(cont *engine.Container) error {
	log.Debugln("Stop container:", cont.Id)
	return d.client.StopContainer(cont.Id, 10)
}

func (d Docker) Remove(cont *engine.Container) error {
	log.Debugln("Remove container:", cont.Id)
	opts := dockerclient.RemoveContainerOptions{
		ID: cont.Id,
	}
	return d.client.RemoveContainer(opts)
}

func (d Docker) List() (conts []*engine.Container, err error) {
	opts := dockerclient.ListContainersOptions{All: true}
	cs, err := d.client.ListContainers(opts)
	if err != nil {
		return
	}
	for _, c := range cs {
		img := engine.Image{
			Name: strings.Split(c.Image, ":")[0],
		}
		cont := &engine.Container{
			Id:    c.ID,
			Image: img,
			Cmd:   strings.Split(c.Command, ","),
			Name:  strings.TrimPrefix(c.Names[0], "/"),
		}
		conts = append(conts, cont)
	}
	return
}

func (d Docker) GetImg(name string) (img engine.Image, err error) {
	if t := strings.Index(name, ":"); t == -1 {
		name += ":latest"
	}
	opts := dockerclient.ListImagesOptions{All: true}
	imgs, err := d.client.ListImages(opts)
	for _, i := range imgs {
		for _, t := range i.RepoTags {
			if name == t {
				img = engine.Image{Id: i.ID, Name: name}
				log.Debugln("Get image:", img.Name)
				return
			}
		}
	}
	err = errors.New("Image does not exists")
	return
}

func (d Docker) PullImg(name string) (img engine.Image, err error) {
	if t := strings.Index(name, ":"); t == -1 {
		name += ":latest"
	}
	opts := dockerclient.PullImageOptions{
		Repository:   name,
		OutputStream: os.Stdout,
	}
	auth := dockerclient.AuthConfiguration{}
	err = d.client.PullImage(opts, auth)
	return d.GetImg(name)
}

func (d Docker) BuildImg(name, dir string) (img engine.Image, err error) {
	f := path.Join(dir, "Dockerfile")
	if _, err = os.Stat(f); os.IsNotExist(err) {
		return
	}
	log.Infof("Building image %s at %s", name, f)
	opts := dockerclient.BuildImageOptions{
		Name:         name,
		OutputStream: os.Stdout,
		ContextDir:   dir,
	}
	if err = d.client.BuildImage(opts); err != nil {
		return
	}
	img = engine.Image{Name: name, Manual: true}
	return
}

func (d Docker) RemoveImg(name string) (err error) {
	return d.client.RemoveImage(name)
}

func New(endpoint, certPath string) (d Docker, err error) {
	var c *dockerclient.Client
	if certPath != "" {
		cert := fmt.Sprintf("%s/cert.pem", certPath)
		key := fmt.Sprintf("%s/key.pem", certPath)
		ca := fmt.Sprintf("%s/ca.pem", certPath)
		c, err = dockerclient.NewTLSClient(endpoint, cert, key, ca)
	} else {
		c, err = dockerclient.NewClient(endpoint)
	}
	if err != nil {
		return
	}
	log.Debugln("Connecting to Docker on:", endpoint)
	if _, err := c.Info(); err != nil {
		return d, err
	}
	d = Docker{client: c}
	return
}
