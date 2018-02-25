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

// Initialise configures the Self node using STUN to attempt to determine the external IP address.
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
