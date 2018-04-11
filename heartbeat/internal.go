package heartbeat

import (
	"fmt"
	"time"
)

// RoutineNormal heartbeats are sent internally to indicate normal status. It satisfies the error interface so we can
// send either a wrapped error or no-error as required.
type RoutineNormal struct {
	origin    string
	Timestamp time.Time
}

// NewRoutineNormal generates a normal heartbeat.
func NewRoutineNormal() RoutineNormal {
	return RoutineNormal{Timestamp: time.Now()}
}

func (n RoutineNormal) Error() string {
	return fmt.Sprintf("Routine Normal (%v)", n.origin)
}

func (n *RoutineNormal) SetOrigin(o string) RoutineNormal {
	n.origin = o
	return *n
}
