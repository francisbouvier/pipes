package utils

import (
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

func download(p string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	log.Debugln("Download status code:", resp.StatusCode)
	if resp.StatusCode == 404 {
		return errors.New("Does not exist")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(p, body, 0755)
}

func getDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	dir := path.Join(u.HomeDir, ".pipes", "tools")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}
	}
	return dir, nil
}

func GetTool(name, url string) (p string, err error) {
	dir, err := getDir()
	if err != nil {
		return
	}
	p = path.Join(dir, name)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		log.Infof("Downloading %s from %s ...", name, url)
		err = download(p, url)
		return p, err
	}
	return
}

func PickServer(pool []string) int {
	// Chose randomly
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	i := r.Intn(len(pool))
	if i > (len(pool) - 1) {
		i -= 1
	}
	return i
}

func AddrToIP(addr string) string {
	// Remove protocol
	ip := strings.TrimPrefix(addr, "tcp://")
	// Remove port
	ip = strings.Split(ip, ":")[0]
	return ip
}

func SplitAddr(addr string) string {
	// Remove multiple
	addr = strings.Split(addr, ",")[0]
	// Remove protocol
	addr = strings.TrimPrefix(addr, "http://")
	return addr
}
