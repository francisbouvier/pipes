package orch

import "github.com/francisbouvier/pipes/src/engine"

type Orch interface {
	Initialize([]string) error
	Join(string) error
	Leave(string) error
	engine.Engine
}
