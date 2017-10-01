package consensus

import (
	"github.com/alowde/dpoller/alert"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/url"
	"time"
)

func Init(in chan url.Status) (routineStatus chan error, err error) {
	routineStatus = make(chan error, 10)
	go checkConsensus(in, routineStatus)

	return routineStatus, nil
}

func checkConsensus(in chan url.Status, routineStatus chan error) {
	for {
		var urlStatuses url.Statuses
		interval := time.After(60 * time.Second)
	timer:
		for {
			select {
			case <-interval:
				if heartbeat.Self.Coordinator { // Only the coordinator checks URL statuses
					deduped := urlStatuses.Dedupe()
					alert.ProcessAlerts(deduped.GetFailed())
				}
				routineStatus <- heartbeat.RoutineNormal{time.Now()}
				break timer
			case s := <-in:
				urlStatuses = append(urlStatuses, s)
			}
		}
	}
}
