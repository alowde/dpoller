package heartbeat

import (
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/logger"
	"github.com/alowde/dpoller/node"
	"time"
)

var log *logrus.Entry

// Initialise configures the logging level for the heartbeat module.
func Initialise(ll logrus.Level) {
	log = logger.New("heartbeat", ll)
}

// Beat holds node status at a point in time, including Coordinator/Feasible Coordinator status.
type Beat struct {
	node.Node
	Coordinator bool
	Feasible    bool
	Timestamp   time.Time
}

// NewBeat returns an initialised Beat.
func NewBeat() Beat {
	return Beat{
		node.Self,
		Self.Coordinator,
		Self.Feasible,
		time.Now(),
	}
}

// Self stores this node's base external heartbeat. It includes the node information as well as whether it seeks to
// hold Coordinator or Feasible Coordinator position.
var Self Beat
