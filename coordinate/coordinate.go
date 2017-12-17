package coordinate

import (
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/node"
	"time"
)

var knownBeats heartbeat.BeatMap
var log *logrus.Entry

func Init(in chan heartbeat.Beat, ll logrus.Level) (statusReport chan error, err error) {

	logrus.SetLevel(ll)
	log = logrus.WithFields(logrus.Fields{
		"routine": "coordinator",
		"ID":      node.Self.ID,
	})

	statusReport = make(chan error, 10)
	knownBeats = heartbeat.NewBeatMap()
	go updateKnownBeats(in, statusReport)

	return statusReport, nil
}

// updateKnownBeats receives heartbeats from the listener and stores the most recent per-node in the knownBeats map
func updateKnownBeats(in chan heartbeat.Beat, statusReport chan error) {

timer:
	for {
		interval := time.After(30 * time.Second)
		select {
		case <-interval:
			// when interval expires, delete beats of nodes that haven't been seen for 120 seconds and evaluate
			log.WithField("nodes", knownBeats.GetNodes()).Info("Aging out nodes")
			knownBeats.AgeOut()
			log.WithField("nodes", knownBeats.GetNodes()).Info("Evaluating nodes")
			knownBeats.Evaluate() // because this is a blocking call we don't need to lock the map
			log.Infoln("Reporting status")
			statusReport <- heartbeat.RoutineNormal{Timestamp: time.Now()}
			continue timer
		case b := <-in:
			knownBeats[b.ID] = b
		}
	}
}
