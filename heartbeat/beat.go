package heartbeat

import (
	"github.com/alowde/dpoller/node"
	"time"
)

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
		coordinator,
		feasibleCoordinator,
		time.Now(),
	}
}

// Self stores this node's base external heartbeat. It includes the node information as well as whether it seeks to
// hold Coordinator or Feasible Coordinator position.
// var Self Beat

var coordinator bool

// GetCoordinator gets current coordinator status
func GetCoordinator() bool {
	return coordinator
}

// SetCoordinator sets current coordinator status
func SetCoordinator(b bool) {
	coordinator = b
}

var feasibleCoordinator bool

// GetFeasibleCoordinator gets current feasible coordinator status
func GetFeasibleCoordinator() bool {
	return feasibleCoordinator
}

// SetFeasibleCoordinator sets current feasible coordinator status
func SetFeasibleCoordinator(b bool) {
	feasibleCoordinator = b
}
