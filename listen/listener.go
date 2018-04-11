// Package listen provides a generic interface for receiving check results and heartbeat messages from other nodes.
// It's expected that a listener will receive from all other nodes, though it may not connect to every one.
// Listeners are one of four routines that must send a heartbeat for the node to be considered healthy. This allows for
// broker/external connectivity checks to be easily incorporated into the node's self-check process.
package listen

import (
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/url/check"
)

// A Listener can connect to some form of external message broker and return messages on the provided
// channels.
type Listener interface {
	Initialise(config string, level logrus.Level) (result chan error, hchan chan heartbeat.Beat, schan chan check.Status, err error)
}
