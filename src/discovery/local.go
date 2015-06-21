package discovery

import (
	"io/ioutil"
	"os/user"
	"path"

	log "github.com/Sirupsen/logrus"
)

type Local struct {
}

func getPath(p string) (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return path.Join(u.HomeDir, ".pipes", p), nil
}

func (l Local) Connect(p string) (data []byte, err error) {
	absP, err := getPath(p)
	if err != nil {
		return
	}
	data, err = ioutil.ReadFile(absP)
	if err != nil {
		log.Debugln("Create a conf")
		err = nil
	}
	return data, nil
}

func (l Local) Save(p string, data []byte) (err error) {
	absP, err := getPath(p)
	if err != nil {
		return
	}
	ioutil.WriteFile(absP, data, 0644)
	log.Infoln("Save conf")
	return nil
}
