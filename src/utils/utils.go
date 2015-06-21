package utils

import (
	"math/rand"
	"strings"
	"time"
)

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
