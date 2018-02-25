package coordinate

import (
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/logger"
	"time"
)

var knownBeats heartbeat.BeatMap

var log *logrus.Entry

// Initialise starts the consensus-checking routine and returns a status channel.
func Initialise(in chan heartbeat.Beat, ll logrus.Level) (statusReport chan error, err error) {

	log = logger.New("coordinate", ll)

	statusReport = make(chan error, 10)
	knownBeats = heartbeat.NewBeatMap()
	go updateKnownBeats(in, statusReport)

	return statusReport, nil
}

// updateKnownBeats receives heartbeats from the listener and stores the most recent per-node in the knownBeats map.
func updateKnownBeats(in chan heartbeat.Beat, statusReport chan error) {

timer:
	for {
		var interval time.Duration
		if heartbeat.Self.Coordinator || heartbeat.Self.Feasible {
			interval = 5 * time.Second
		} else {
			interval = 30 * time.Second
		}
		coordTimer := time.After(interval)
		for {
			select {
			case <-coordTimer:
				log.Debug("interval expired")
				// when interval expires, delete beats of nodes that haven't been seen for 21 seconds and evaluate
				log.WithField("nodes", knownBeats.GetNodes()).Debug("Aging out nodes")
				knownBeats.AgeOut()
				log.WithField("nodes", knownBeats.GetNodes()).Debug("Evaluating nodes")
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
}
