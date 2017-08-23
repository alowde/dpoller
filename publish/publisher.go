package publish

import (
	//	"github.com/alowde/dpoller/heartbeat"
	//"github.com/alowde/dpoller/url"
	"time"
)

// A Publisher accepts heartbeat and status messages and sends them to other nodes
type Publisher interface {
	Init(config string) (err error)
	Publish(i interface{}, deadline <-chan time.Time) (err error)
}
