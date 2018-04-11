// Package publish provides a generic interface for sending check result and heartbeat messages to other nodes.
// It's expected that a publisher will publish to all other nodes, though it may not connect to every one.
package publish

import (
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/url/check"
	"time"
)

// A Publisher accepts heartbeat and status messages and sends them to other nodes.
type Publisher interface {
	Init(string, chan heartbeat.Beat, chan check.Status, logrus.Level) (err error)
	Publish(i interface{}, deadline <-chan time.Time) (err error)
}
