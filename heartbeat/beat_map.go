package heartbeat

import "time"

// BeatMap is used in various cases where it's convenient to associate []Beat to a given node.
type BeatMap map[int64]Beat

func NewBeatMap() BeatMap {
	return make(map[int64]Beat)
}

// AgeOut removes beats that have not been seen in the last 21 seconds. This is the time required for a node to be
// unresponsive before it will be ignored.
// The time is selected as a compromise between the risk of creating network partitions and the risk of missing
// failed tests that require alerting.
func (bm BeatMap) AgeOut() {
	for k, v := range bm {
		if time.Now().Sub(v.Timestamp) > 21*time.Second {
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
