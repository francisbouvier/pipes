package discovery

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/francisbouvier/pipes/src/orch"
	"github.com/francisbouvier/pipes/src/orch/swarm"
	"github.com/francisbouvier/pipes/src/store"
)

type Discovery interface {
	Connect(string) ([]byte, error)
	Save(string, []byte) error
}

func Initialize(c *cli.Context) (err error) {

	// Conf
	name := c.String("name")
	cf, err := getConf(c)
	if err != nil {
		return
	}
	if _, err := cf.GetPool(name); err == nil {
		return errors.New("Pool already exist")
	}

	// Servers
	servers := c.StringSlice("servers")
	if len(servers) == 0 {
		dockerHost := os.Getenv("DOCKER_HOST")
		if dockerHost == "" {
			dockerHost = "unix:///var/run/docker.sock"
		}
		servers = append(servers, dockerHost)
	} else if len(servers) == 1 {
		serversArray := strings.SplitN(servers[0], ",", -1)
		servers = []string{}
		for _, server := range serversArray {
			serverFullAddr := fmt.Sprintf("tcp://%s:2375", server)
			servers = append(servers, serverFullAddr)
		}
	}

	// Store
	st, err := store.Get("etcd")
	if err != nil {
		return err
	}
	if err = st.Initialize(name, servers); err != nil {
		return err
	}

	// Orch
	var orchest orch.Orch
	orchest = swarm.Swarm{Store: st}
	if err = orchest.Initialize(servers); err != nil {
		return err
	}

	// Wamp Router
	swarm, err := swarm.New(st)
	if err != nil {
		return err
	}
	if _, err = wampRouter(swarm); err != nil {
		return err
	}

	cf.SetPool(name, st.Addr())
	cf.SetMainPool(name)
	if err = cf.Save(); err != nil {
		return err
	}
	fmt.Println("Cluster initialized")
	return nil
}

func GetStore(c *cli.Context) (st store.Store, err error) {
	cf, err := getConf(c)
	if err != nil {
		return
	}
	_, addr, err := cf.GetMainPool()
	if err != nil {
		return
	}
	st, err = store.Get("etcd")
	if err != nil {
		return
	}
	st.New(addr)
	return
}
