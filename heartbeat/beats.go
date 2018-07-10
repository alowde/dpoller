package heartbeat

import (
	"fmt"
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

func (beats Beats) checkIdExists(nodeID int64) bool {
	for _, b := range beats {
		if b.ID == nodeID {
			return true
		}
	}
	return false
}

// Evaluate assesses the set of known nodes to determine which node has/should have the Coordinator role. It implements
// the following decision tree:
// - If there's no coordinator and this node is the best feasible coordinator, take the role
// - If there's one or more coordinators and this node is one of them:
// -- If this node is the best coordinator no action is required
// -- If this node is not the best coordinator, reset to no roles
// - If there's no feasible coordinator and this node is the best feasible coordinator, take the role
// - If there's one or more feasible coordinators and this node is one of them:
// -- If this node is the best feasible coordinator no action is required
// -- If this node is not the best feasible coordinator reset to no roles
// - Finally, if there's a coordinator and feasible coordinator but we're not them, no action is required.
// Implementing this as a two-phase selection is intended to keep the inevitable flapping due to unreliable networks at
// the first phase. If a coordinator loses its position it won't immediately compete for it again which should reduce
// the rate of role-change.
// It's possible that the evaluation algorithm could be modified to take into account perceived connection stability
// (perhaps based on number of known nodes) but the current one has the advantage of being very simple to reason about.
func (beats Beats) Evaluate(isCoord, isFeas bool, nodeID int64) (shouldBeCoordinator, shouldBeFeasible bool) {

	// Beats must include the given nodeID or we can't continue - enforcing this constraint simplifies this method and
	// puts responsibility back on the caller to provide sane data
	// TODO: Extend the Evaluate() method to throw an error instead of panicking
	if !beats.checkIdExists(nodeID) {
		log.WithField("beats", beats).
			Fatal("Can't evaluate beats without self included")
	}

	bf, _ := beats.bestFeas()

	// First check if we're a feasible coordinator that should be promoted
	if beats.CoordCount() == 0 {
		log.Infoln("No coordinators exist at all")

		if isFeas && bf == nodeID {
			log.Infoln("This node is the best feasible coordinator, promote")
			shouldBeCoordinator = true
			shouldBeFeasible = false
			return
		}
	}

	// Check if we're a competing coordinator. If so and we're not the best unset both roles and return, otherwise set
	// the coordinator role and return.
	if beats.CoordCount() > 0 && isCoord {
		if bac, _ := beats.bestActiveCoord(); bac != nodeID {
			shouldBeCoordinator = false
			shouldBeFeasible = false
			return
		}
		shouldBeCoordinator = true
		shouldBeFeasible = false
		return
	}

	// check if we need to take the feasible coordinator role
	if beats.FeasCount() == 0 && bf == nodeID {
		shouldBeCoordinator = false
		shouldBeFeasible = true
		return
	}

	// Check if we're a competing feasible coordinator. If so and we're not the best unset both roles and return,
	// otherwise set the feasible coordinator role and return.
	if beats.FeasCount() > 0 && isFeas {
		if baf, _ := beats.bestActiveFeas(); baf != nodeID {
			shouldBeCoordinator = false
			shouldBeFeasible = false
			return
		}
		shouldBeCoordinator = false
		shouldBeFeasible = true
		return
	}

	// No action required - there's a coordinator and feasible coordinator but we're not them.
	shouldBeCoordinator = false
	shouldBeFeasible = false
	return
}
