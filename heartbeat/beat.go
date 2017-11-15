package heartbeat

import "fmt"
import "math"
import "github.com/alowde/dpoller/node"
import "time"

// RoutineNormal is used by routines to indicate normal healthy status
type RoutineNormal struct {
	origin    string
	Timestamp time.Time
}

func (n RoutineNormal) Error() string {
	return fmt.Sprintf("Routine Normal (%v)", n.origin)
}

func (n *RoutineNormal) SetOrigin(o string) RoutineNormal {
	n.origin = o
	return *n
}

func NewRoutineNormal() RoutineNormal {
	return RoutineNormal{Timestamp: time.Now()}
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

func (beats Beats) bestActiveFeas() (feasID int64, e error) {
	e = fmt.Errorf("No values")
	feasID = math.MaxInt64
	for _, b := range beats {
		if (b.ID < feasID) && b.Feasible {
			e = nil
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

// Evaluate assesses the set of known nodes to determine which node has/should have the Coordinator role
func (beats Beats) Evaluate() {
	if beats.coordCount() == 0 {
		if Self.Coordinator { // This node is the uncontested coordinator, no further evaluation
			Self.Feasible = false
			return
		} else { // No coordinators exist at all
			if Self.Feasible { // This node is the feasible coordinator, promote
				Self.Coordinator = true
				Self.Feasible = false
				return
			} else { // This node is not the feasible coordinator, assess feasible coordinator status
				beats.evaluateFeas()
				return
			}
		}
	} else {
		if Self.Coordinator { // This node is a contested coordinator
			best, _ := beats.bestActiveCoord()
			if best > Self.ID { // This node is not the best coordinator, update Self and assess feasible coordinators
				Self.Coordinator = false
				beats.evaluateFeas()
				return
			} else { // This node is the best coordinator, cease further evaluation
				Self.Feasible = false
				return
			}
		} else { // There is one or more coordinators already, move to evaluating feasible coordinators
			beats.evaluateFeas()
			return
		}
	}
}

// Evaluate assesses the set of known nodes to determine which node has/should have the Feasible Coordinator role
func (beats Beats) evaluateFeas() {
	if beats.feasCount() == 0 {
		if Self.Feasible { // This node is the uncontested feasible coordinator, no further evaluation
			return
		} else { // No feasible coordinators exist
			best, _ := beats.bestFeas()
			if best < Self.ID { // This node is the best possible feasible coordinator, set
				Self.Feasible = true
			}
			return
		}
	} else {
		if Self.Feasible { // This node is a contested feasible coordinator
			best, _ := beats.bestActiveFeas()
			if best > Self.ID { // This node is not the best feasible coordinator in contention, unset
				// WARNING: This means incomplete/one-way message transmission may cause flapping!
				Self.Feasible = false
			}
		}
	}
}

type BeatMap map[int64]Beat

func NewBeatMap() BeatMap {
	return make(map[int64]Beat)
}

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
