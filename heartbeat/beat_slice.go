package heartbeat

import (
	"fmt"
	"math"
)

// Beats is a slice of Beat, mainly used for convenient aggregate calculation functions.
type Beats []Beat

func (beats Beats) CoordCount() (count int) {
	for _, b := range beats {
		if b.Coordinator {
			count++
		}
	}
	return
}

func (beats Beats) FeasCount() (count int) {
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
