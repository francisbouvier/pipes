package etcd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	client "github.com/coreos/go-etcd/etcd"
	"github.com/francisbouvier/pipes/src/engine"
	"github.com/francisbouvier/pipes/src/engine/docker"
	"github.com/francisbouvier/pipes/src/store"
)

const (
	LOCALHOST = "127.0.0.1"
	IMAGE     = "quay.io/coreos/etcd"
)

var QUORUM = 3

type Node struct {
	// id     string
	addr   string
	status bool
}

type Etcd struct {
	nodes  []*Node
	addr   string
	client *client.Client
}

func init() {
	store.Register("etcd", &Etcd{})
}

func getNodes(addr string) (nodes []*Node, addrs []string) {
	for _, nodeAddr := range strings.Split(addr, ",") {
		node := &Node{addr: nodeAddr}
		nodes = append(nodes, node)
		addrs = append(addrs, nodeAddr)
	}
	return
}

func (st *Etcd) Initialize(token string, servers []string) error {
	log.Debugln("Installing etcd store ...")

	// Servers and IPs
	ips := []string{}
	for _, server := range servers {
		list := strings.Split(server, "://")
		var ip string
		if list[0] == "unix" {
			ip = LOCALHOST
		} else {
			ip = strings.Split(list[1], ":")[0]
		}
		ips = append(ips, "http://"+ip)
	}
	// Warning if there is not enough server to reach the QUORUM
	if len(ips) < QUORUM {
		log.Infof("Not enough servers to reach the Quorum (%d)", QUORUM)
	}

	// Engines
	engines := []engine.Engine{}
	for _, server := range servers {
		eng, err := docker.New(server, "")
		if err != nil {
			return err
		}
		engines = append(engines, eng)
	}

	// Image
	img := engine.Image{Name: IMAGE}

	// Containers
	localIP := "http://0.0.0.0"
	containers := []*engine.Container{}
	addr := ""
	clusterAddr := ""
	for i, ip := range ips {
		name := "etcd" + strconv.Itoa(i+1)
		cl := strconv.Itoa(4100 + i + 1)
		peer := strconv.Itoa(7100 + i + 1)
		cmd := []string{
			"-name", name,
			"-initial-cluster-token", token,
			"-initial-cluster-state", "new",
			"-listen-client-urls", fmt.Sprintf("%s:%s", localIP, cl),
			"-listen-peer-urls", fmt.Sprintf("%s:%s", localIP, peer),
			"-initial-advertise-peer-urls", fmt.Sprintf("%s:%s", ip, peer),
			"-advertise-client-urls", fmt.Sprintf("%s:%s", ip, cl),
		}
		clusterAddr += fmt.Sprintf("%s=%s:%s", name, ip, peer)
		addr += fmt.Sprintf("%s:%s", ip, cl)
		if i < (len(ips) - 1) {
			addr += ","
			clusterAddr += ","
		}
		container := &engine.Container{
			Name:     name,
			Hostname: name,
			Image:    img,
			Ports: []map[string]string{
				map[string]string{cl: cl},
				map[string]string{peer: peer},
			},
			Cmd: cmd,
		}
		// If ruuning in localhost we have use "host" network mode
		// in order to allow each etcd peer to see each other
		if ip == "http://"+LOCALHOST {
			container.NetworkMode = "host"
		}
		containers = append(containers, container)
	}
	for i, _ := range ips {
		container := containers[i]
		container.Cmd = append(container.Cmd, "-initial-cluster", clusterAddr)
		log.Debugln("Getting img")
		if _, err := engines[i].GetImg(img.Name); err != nil {
			if _, err = engines[i].PullImg(img.Name); err != nil {
				return err
			}
		}
		err := engines[i].Run(container)
		if err != nil {
			return err
		}
	}

	// Connect
	log.Debugln("Connecting to store on:", addr)
	st.addr = addr
	var addrs []string
	st.nodes, addrs = getNodes(addr)
	st.client = client.NewClient(addrs)
	// TODO: check cluster health instead of dirty sleep
	time.Sleep(500 * time.Millisecond)
	fmt.Println("Etcd store running on:", st.Addr())
	return nil
}

func (st Etcd) get(key, dir string) (resp *client.Response, err error) {
	if dir != "" {
		key = fmt.Sprintf("/%s/%s", dir, key)
	}
	resp, err = st.client.Get(key, false, false)
	return
}

func (st Etcd) Read(key, dir string) (value string, err error) {
	resp, err := st.get(key, dir)
	if err != nil {
		return
	}
	value = resp.Node.Value
	return
}

func (st Etcd) List(key, dir string) (keys []string, err error) {
	resp, err := st.get(key, dir)
	if err != nil {
		return
	}
	for _, elem := range resp.Node.Nodes {
		prefix := fmt.Sprintf("/%s/", key)
		if dir != "" {
			prefix = fmt.Sprintf("/%s/%s/", dir, key)
		}
		keys = append(keys, strings.TrimPrefix(elem.Key, prefix))
	}
	return
}

func (st Etcd) Write(key, value, dir string) (err error) {
	if key == "" && dir == "" {
		return errors.New("You need to provide at least either key or dir")
	}
	if key == "" {
		key = dir
	} else if dir != "" {
		key = fmt.Sprintf("/%s/%s", dir, key)
	}
	if value != "" {
		_, err = st.client.Set(key, value, 0)
	} else {
		_, err = st.client.CreateDir(key, 0)
	}
	return
}

func (st Etcd) Delete(key, dir string) (err error) {
	if dir != "" {
		key = fmt.Sprintf("/%s/%s", dir, key)
	}
	_, err = st.client.Delete(key, true)
	return
}

func (st Etcd) Addr() (addr string) {
	return st.addr
}
