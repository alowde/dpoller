package consensus

import (
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/alert"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/logger"
	"github.com/alowde/dpoller/url"
	"github.com/alowde/dpoller/url/check"
	"time"
)

var log *logrus.Entry

// Initialise starts the consensus-checking routine and returns a status channel.
func Initialise(in chan check.Status, ll logrus.Level) (routineStatus chan error, err error) {

	log = logger.New("consensus", ll)

	routineStatus = make(chan error, 10)
	go checkConsensus(in, routineStatus)

	return routineStatus, nil
}

func checkConsensus(in chan check.Status, routineStatus chan error) {
	for {
		var urlStatuses check.Statuses
		interval := time.After(60 * time.Second)
	collectLoop:
		for {
			select {
			case <-interval:
				if heartbeat.Self.Coordinator { // Only the coordinator checks URL statuses
					dd := urlStatuses.Dedupe()
					statusSet := dd.PerCheckName() // Dedupe and group status check results by name
					log.WithField("status count", len(statusSet)).
						Info("checking consensus")
					for n, statuses := range statusSet { // Calculate the aggregate statistics for each set of checks
						if r, err := statuses.CalculateResult(); err == nil { // Ignore empty statussets
							c, _ := url.Checks.ByName(n)          // Get an absolute copy of the check configuration
							if r.PassPercent < c.AlertThreshold { // If the result PassPercent is lower than the configd
								log.WithField("check name", c.Name).
									WithField("alert threshold", r.PassPercent).
									WithField("passed checks", r.PassPercent).
									Debug("alerting on failed check")
								alert.ProcessAlerts(c, r) // check threshold, send an alert
							}
						}
					}
				}
				routineStatus <- heartbeat.RoutineNormal{Timestamp: time.Now()}
				break collectLoop
			case s := <-in:
				urlStatuses = append(urlStatuses, s)
			}
		}
	}
}
