package heartbeat

import (
	"fmt"
	"github.com/alowde/dpoller/node"
	"math"
)

// Beats is a slice of Beat, mainly used for convenient aggregate calculation functions.
type Beats []Beat

// CoordCount calculates the number of nodes in a Beats array that are Coordinators
func (beats Beats) CoordCount() (count int) {
	for _, b := range beats {
		if b.Coordinator {
			count++
		}
	}
	return
}

// FeasCount calculates the number of nodes in a Beats array that are Feasible Coordinators
func (beats Beats) FeasCount() (count int) {
	for _, b := range beats {
		if b.Feasible {
			count++
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

// TODO: refactor these routines to have no knowledge of the node's state

// Evaluate assesses the set of known nodes to determine which node has/should have the Coordinator role
func (beats Beats) Evaluate() {
	if beats.CoordCount() == 0 {
		if Self.Coordinator {
			log.Infoln("This node is the uncontested coordinator, no further evaluation")
			Self.Feasible = false
			return
		}
		log.Infoln("No coordinators exist at all")
		if Self.Feasible {
			log.Infoln("This node is the feasible coordinator, promote")
			Self.Coordinator = true
			Self.Feasible = false
			return
		}
		log.Infoln("This node is not the feasible coordinator, assess feasible coordinator status")
		beats.evaluateFeas()
		return
	}
	if Self.Coordinator {
		best, _ := beats.bestActiveCoord()
		if best < node.Self.ID {
			log.Infoln("This node is not the best coordinator, update Self and assess feasible coordinators")
			Self.Coordinator = false
			beats.evaluateFeas()
			return
		}
		log.Infoln("node is the best coordinator, cease further evaluation")
		Self.Feasible = false
		return

	}
	log.Infoln("There is one or more coordinators already, move to evaluating feasible coordinators")
	beats.evaluateFeas()
}

// Evaluate assesses the set of known nodes to determine which node has/should have the Feasible Coordinator role
func (beats Beats) evaluateFeas() {
	if beats.FeasCount() == 0 {
		if Self.Feasible {
			log.Infoln("This node is the uncontested feasible coordinator, no further evaluation")
			return
		}
		log.Infoln("No feasible coordinators exist")
		best, _ := beats.bestFeas()
		if best < node.Self.ID {
			log.Infoln("This node is not the best possible feasible coordinator, unset")
			Self.Feasible = false
			return

		}
		log.Infoln("This node is the best possible feasible coordinator, set")
		Self.Feasible = true
		return
	}
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
