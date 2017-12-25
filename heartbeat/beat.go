package heartbeat

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/node"
	"math"
	"time"
)

var log *logrus.Entry

func Init(ll logrus.Level) {

	logrus.SetLevel(ll)
	log = logrus.WithFields(logrus.Fields{
		"routine": "heartbeat",
		"ID":      node.Self.ID,
	})
}

// RoutineNormal is used by routines to indicate normal healthy status
type RoutineNormal struct {
	origin    string
	Timestamp time.Time
}

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

type Beat struct {
	node.Node
	Coordinator bool
	Feasible    bool
	Timestamp   time.Time
}

func NewBeat() Beat {
	return Beat{
		node.Self,
		Self.Coordinator,
		Self.Feasible,
		time.Now(),
	}
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
	e = fmt.Errorf("no values")
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
	e = fmt.Errorf("no values")
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
	e = fmt.Errorf("no values")
	feasID = math.MaxInt64
	for _, b := range beats {
		e = nil
		if b.ID < feasID && !b.Coordinator {
			fmt.Printf("found lower beat ID %v\n", b.ID)
			feasID = b.ID
		}
	}
	return
}

func (beats Beats) bestActiveFeas() (feasID int64, e error) {
	e = fmt.Errorf("no values")
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
		if Self.Coordinator {
			log.Infoln("This node is the uncontested coordinator, no further evaluation")
			Self.Feasible = false
			return
		} else {
			log.Infoln("No coordinators exist at all")
			if Self.Feasible {
				log.Infoln("This node is the feasible coordinator, promote")
				Self.Coordinator = true
				Self.Feasible = false
				return
			} else {
				log.Infoln("This node is not the feasible coordinator, assess feasible coordinator status")
				beats.evaluateFeas()
				return
			}
		}
	} else {
		if Self.Coordinator {
			best, _ := beats.bestActiveCoord()
			if best < node.Self.ID {
				log.Infoln("This node is not the best coordinator, update Self and assess feasible coordinators")
				Self.Coordinator = false
				beats.evaluateFeas()
				return
			} else {
				log.Infoln("node is the best coordinator, cease further evaluation")
				Self.Feasible = false
				return
			}
		} else {
			log.Infoln("There is one or more coordinators already, move to evaluating feasible coordinators")
			beats.evaluateFeas()
			return
		}
	}
}

// Evaluate assesses the set of known nodes to determine which node has/should have the Feasible Coordinator role
func (beats Beats) evaluateFeas() {
	if beats.feasCount() == 0 {
		if Self.Feasible {
			log.Infoln("This node is the uncontested feasible coordinator, no further evaluation")
			return
		} else {
			log.Infoln("No feasible coordinators exist")
			best, _ := beats.bestFeas()
			if best < node.Self.ID {
				log.Infoln("This node is not the best possible feasible coordinator, unset")
				Self.Feasible = false
				return

			} else {
				log.Infoln("This node is the best possible feasible coordinator, set")
				Self.Feasible = true
				return
			}

		}
	} else {
		if Self.Feasible {
			log.Infoln("This node is a contested feasible coordinator")
			best, _ := beats.bestActiveFeas()
			if best < node.Self.ID {
				log.Infoln("This node is not the best feasible coordinator in contention, unset")
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

func (bm BeatMap) ToBeats() (b Beats) {
	for _, v := range bm {
		b = append(b, v)
	}
	return
}

func (bm BeatMap) Evaluate() {
	ba := bm.ToBeats()
	ba.Evaluate()
}

func (bm BeatMap) GetNodes() (n []int64) {
	for k := range bm {
		n = append(n, k)
	}
	return
}
