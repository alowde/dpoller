package coordinate

import (
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/node"

	"github.com/mattn/go-colorable"
	"time"
)

var knownBeats heartbeat.BeatMap

var log *logrus.Entry

func Init(in chan heartbeat.Beat, ll logrus.Level) (statusReport chan error, err error) {

	var logger = logrus.New()
	logger.Formatter = &logrus.TextFormatter{ForceColors: true}
	logger.Out = colorable.NewColorableStdout()
	logger.SetLevel(ll)

	log = logger.WithFields(logrus.Fields{
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
			log.Debug("interval expired")

			// when interval expires, delete beats of nodes that haven't been seen for 120 seconds and evaluate
			//log.WithField("nodes", knownBeats.GetNodes()).Debug("Aging out nodes")
			knownBeats.AgeOut()
			//log.WithField("nodes", knownBeats.GetNodes()).Debug("Evaluating nodes")
			knownBeats.Evaluate() // because this is a blocking call we don't need to lock the map
			log.WithFields(logrus.Fields{
				"coordinators":         knownBeats.ToBeats().CoordCount(),
				"feasibleCoordinators": knownBeats.ToBeats().FeasCount(),
				"is_coordinator":       heartbeat.Self.Coordinator,
				"is_feasible":          heartbeat.Self.Feasible,
			}).Info("Finished evaluating feasible/coordinators")
			statusReport <- heartbeat.RoutineNormal{Timestamp: time.Now()}
			continue timer
		case b := <-in:
			log.Debug("beat in")
			knownBeats[b.ID] = b
		}
	}
}
