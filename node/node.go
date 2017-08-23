package node

import (
	"fmt"
	"github.com/ccding/go-stun/stun"
	"github.com/pkg/errors"
	"math/rand"
	"net"
	"time"
)

type Node struct {
	ID   int64
	EIP  net.IP
	Name string
}

var Self Node

func Initialise() error {
	_, host, err := stun.NewClient().Discover()
	if err != nil {
		return errors.Wrap(err, "failed to discover external IP address")
	}
	ip, _, err := net.ParseCIDR(fmt.Sprintf("%v/32", host.IP()))
	if err != nil {
		return errors.Wrapf(err, "STUN returned unparseable address %v", host.IP())
	}

	rand.Seed(time.Now().UnixNano())
	Self.ID = rand.Int63()
	Self.EIP = ip
	return nil
}
