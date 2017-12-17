package listen

import (
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/url/urltest"
)

// A Listener can connect to some form of external message broker and return messages on the provided
// channels
type Listener interface {
	Init(config string, level logrus.Level) (result chan error, hchan chan heartbeat.Beat, schan chan urltest.Status, err error)
}
