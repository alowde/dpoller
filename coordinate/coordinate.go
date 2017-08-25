package coordinate

import (
	"github.com/alowde/dpoller/heartbeat"
	"time"
)

var knownBeats heartbeat.BeatMap

func Init(in chan heartbeat.Beat) (statusReport chan error, err error) {
	statusReport = make(chan error, 10)
	go updateKnownBeats(in, statusReport)

	return statusReport, nil
}

// updateKnownBeats receives heartbeats from the listener and stores the most recent per-node in the knownBeats map
func updateKnownBeats(in chan heartbeat.Beat, statusReport chan error) {
	for {
		interval := time.After(10 * time.Second)
	timer:
		for {
			select {
			case b := <-in:
				knownBeats[b.ID] = b
			case <-interval:
				// when interval expires, delete beats of nodes that haven't been seen for 120 seconds and evaluate
				knownBeats.AgeOut()
				evaluateKnownBeats() // because this is a blocking call we don't need to lock the map
				statusReport <- heartbeat.RoutineNormal{time.Now()}
				break timer
			}
		}
	}
}

func evaluateKnownBeats() {

}
