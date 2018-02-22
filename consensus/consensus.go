package consensus

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/alert"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/logger"
	"github.com/alowde/dpoller/url/check"
	"time"
)

var log *logrus.Entry

func Init(in chan check.Status, ll logrus.Level) (routineStatus chan error, err error) {

	log = logger.New("consensus", ll)

	routineStatus = make(chan error, 10)
	go checkConsensus(in, routineStatus)

	return routineStatus, nil
}

func checkConsensus(in chan check.Status, routineStatus chan error) {
	for {
		var urlStatuses check.Statuses
		interval := time.After(60 * time.Second)
	timer:
		for {
			select {
			case <-interval:
				if heartbeat.Self.Coordinator { // Only the coordinator checks URL statuses
					log.Debug("Checking consensus")
					deduped := urlStatuses.Dedupe()
					alert.ProcessAlerts(deduped.GetFailed())
				}
				routineStatus <- heartbeat.RoutineNormal{Timestamp: time.Now()}
				break timer
			case s := <-in:
				urlStatuses = append(urlStatuses, s)
			}
		}
	}
}
