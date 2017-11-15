package coordinate

import (
	log "github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/mattn/go-colorable"
	"time"
)

var knownBeats heartbeat.BeatMap

func Init(in chan heartbeat.Beat) (statusReport chan error, err error) {
	statusReport = make(chan error, 10)
	knownBeats = heartbeat.NewBeatMap()
	go updateKnownBeats(in, statusReport)

	return statusReport, nil
}

// updateKnownBeats receives heartbeats from the listener and stores the most recent per-node in the knownBeats map
func updateKnownBeats(in chan heartbeat.Beat, statusReport chan error) {
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetOutput(colorable.NewColorableStdout())
timer:
	for {
		interval := time.After(10 * time.Second)
		select {
		case <-interval:
			// when interval expires, delete beats of nodes that haven't been seen for 120 seconds and evaluate
			log.Infoln("Aging out nodes")
			knownBeats.AgeOut()
			log.Infoln("Evaluating nodes")
			knownBeats.Evaluate() // because this is a blocking call we don't need to lock the map
			log.Infoln("Reporting status")
			statusReport <- heartbeat.RoutineNormal{Timestamp: time.Now()}
			continue timer
		case b := <-in:
			knownBeats[b.ID] = b
		}
	}
}
