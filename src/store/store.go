package store

import (
	"errors"
	"fmt"
)

type Store interface {
	Initialize(string, []string) error
	New(string)
	Read(string, string) (string, error)
	List(string, string) ([]string, error)
	Write(string, string, string) error
	Delete(string, string) error
	Addr() string
}

var stores map[string]Store

func init() {
	stores = make(map[string]Store)
}

func Register(name string, st Store) {
	stores[name] = st
}

func Get(name string) (Store, error) {
	st, prs := stores[name]
	if !prs {
		msg := fmt.Sprintf("Store not registred: %s", name)
		return nil, errors.New(msg)
	}
	return st, nil
}
