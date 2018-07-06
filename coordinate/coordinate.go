// Package coordinate contains the coordinate routine that's responsible for electing a Coordinator and Feasible
// Coordinator node. Normally the node generates a heartbeat every 30 seconds, but the C and FC nodes generate a
// heartbeat every five seconds to reduce downtime.
// Coordinate is one of four routines that must send a heartbeat for the node to be considered healthy.
package coordinate

import (
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/logger"
	"github.com/alowde/dpoller/node"
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
		if heartbeat.GetCoordinator() || heartbeat.GetFeasibleCoordinator() {
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
				c, f := knownBeats.ToBeats().Evaluate(heartbeat.GetCoordinator(), heartbeat.GetFeasibleCoordinator(), node.Self.ID)
				heartbeat.SetCoordinator(c)
				heartbeat.SetFeasibleCoordinator(f)
				log.WithFields(logrus.Fields{
					"coordinators":         knownBeats.ToBeats().CoordCount(),
					"feasibleCoordinators": knownBeats.ToBeats().FeasCount(),
					"is_coordinator":       heartbeat.GetCoordinator(),
					"is_feasible":          heartbeat.GetFeasibleCoordinator(),
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
