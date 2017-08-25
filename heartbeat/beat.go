package heartbeat

import "fmt"
import "math"
import "github.com/alowde/dpoller/node"
import "time"

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
	Timestamp   time.Time
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

func (beats Beats) bestActiveCoord() (coordID int64, e error) {
	e = fmt.Errorf("No values")
	coordID = math.MaxInt64
	for _, b := range beats {
		e = nil
		if (b.ID < coordID) && b.Coordinator {
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

func (beats Beats) toBeatMap() (result BeatMap) {
	var t BeatMap = make(map[int64]Beat)
	for _, v := range beats {
		if _, ok := t[v.ID]; ok {
			if t[v.ID].Timestamp.Before(v.Timestamp) { // keep only the most recently generated heartbeat for each node
				t[v.ID] = v
			}
		} else {
			t[v.ID] = v
		}
	}
	return t
}

// Evaluate sets various status parameters for the Node based on its neighbours
func (beats Beats) Evaluate() {
	if beats.coordCount() == 0 {
		if Self.Coordinator { // This node is the uncontested coordinator, no further evaluation
			return
		} else { // No coordinators exist at all, assess feasible coordinators
			beats.evaluateFeas()
		}
	} else {
		if Self.Coordinator { // This node is a contested coordinator
			best, _ := beats.bestActiveCoord()
			if best > Self.ID { // This node is not the best coordinator, update Self and assess feasible coordinators
				Self.Coordinator = false
				beats.evaluateFeas()
			} else {
				return // This node is the best coordinator, cease further evaluation
			}
		} else { // There is a single coordinator already, move to evaluating feasible coordinators
			beats.evaluateFeas()
		}
	}
}

func (beats Beats) evaluateFeas() {

}

type BeatMap map[int64]Beat

func (bm BeatMap) AgeOut() {
	for k, v := range bm {
		if time.Now().Sub(v.Timestamp) > 120*time.Second {
			delete(bm, k)
		}
	}
}

func (bm BeatMap) ToBeats() Beats {
	var r Beats
	for _, v := range bm {
		r = append(r, v)
	}
	return r
}

func (bm BeatMap) Evaluate() {
	ba := bm.ToBeats()
	ba.Evaluate()
}

func NewBeat() Beat {
	return Beat{
		node.Self,
		Self.Coordinator,
		Self.Feasible,
		time.Now(),
	}
}
