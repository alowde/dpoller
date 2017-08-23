package heartbeat

import "fmt"
import "math"
import "github.com/alowde/dpoller/node"
import "time"
import "github.com/pkg/errors"

// RoutineNormal is used by routines to indicate normal healthy status
type RoutineNormal struct {
	Timestamp time.Time
}

func (n RoutineNormal) Error() string {
	return ""
}

type Beat struct {
	node.Node
	Coordinator bool
	Feasible    bool
	Timestamp   int64
}

var Self Beat

type Beats []Beat

func (beats Beats) coordCount() (count int) {
	for _, b := range beats {
		if b.Coordinator {
			count++
		}
	}
	return
}

func (beats Beats) feasCount() (count int) {
	for _, b := range beats {
		if b.Feasible {
			count++
		}
	}
	return
}

func (beats Beats) bestCoord() (coordID int64, e error) {
	e = fmt.Errorf("No values")
	coordID = math.MaxInt64
	for _, b := range beats {
		e = nil
		if b.ID < coordID {
			coordID = b.ID
		}
	}
	return
}

func (beats Beats) bestFeas() (feasID int64, e error) {
	e = fmt.Errorf("No values")
	feasID = math.MaxInt64
	for _, b := range beats {
		e = nil
		if b.ID < feasID {
			feasID = b.ID
		}
	}
	return
}

func (beats Beats) dedupe() (result Beats) {
	t := make(map[int64]Beat)
	for _, v := range beats {
		if _, ok := t[v.ID]; ok {
			if t[v.ID].Timestamp < v.Timestamp { // keep only the most recently generated heartbeat for each node
				t[v.ID] = v
			}
		} else {
			t[v.ID] = v
		}
	}
	for _, v := range t {
		result = append(result, v)
	}
	return
}

func NewBeat() Beat {
	return Beat{
		node.Self,
		Self.Coordinator,
		Self.Feasible,
		time.Now().Unix(),
	}
}
