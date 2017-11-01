package publish

import (
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/url/urltest"
	"time"
)

// A Publisher accepts heartbeat and status messages and sends them to other nodes
type Publisher interface {
	Init(string, chan heartbeat.Beat, chan urltest.Status) (err error)
	Publish(i interface{}, deadline <-chan time.Time) (err error)
}
