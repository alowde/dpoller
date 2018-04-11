// Package node provides the Node type and holds the current node's unique identifying information.
package node

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/ccding/go-stun/stun"
	"github.com/mattn/go-colorable"
	"github.com/pkg/errors"
	"math/rand"
	"net"
	"time"
)

// Node is an instance of the dpoller application.
type Node struct {
	ID   int64
	EIP  net.IP
	Name string
}

// Self is the current running node.
var Self Node

var log *logrus.Entry

// Initialise sets this node's unique details: ID from the PRNG, external IP (EIP) from STUN
// 63-bit random UID isn't ideal, but probability of collision is around  2.1e-15 for a 200-node cluster, which
// should be more than sufficient.
func Initialise(l logrus.Level) error {

	var logger = logrus.New()
	logger.Formatter = &logrus.TextFormatter{ForceColors: true}
	logger.Out = colorable.NewColorableStdout()
	logger.SetLevel(l)

	log = logger.WithField("routine", "node")

	log.Debug("Attempting to determine external IP address")
	_, host, err := stun.NewClient().Discover()
	if err != nil {
		return errors.Wrap(err, "failed to discover external IP address")
	}
	ip, _, err := net.ParseCIDR(fmt.Sprintf("%v/32", host.IP()))
	if err != nil {
		return errors.Wrapf(err, "STUN returned unparseable address %v", host.IP())
	}
	log.WithField("ipAddress", ip).Debug("Determined external IP address")

	rand.Seed(time.Now().UnixNano())
	Self.ID = rand.Int63()
	log.WithField("nodeID", Self.ID).Debug("Setting ID")
	Self.EIP = ip
	return nil
}
