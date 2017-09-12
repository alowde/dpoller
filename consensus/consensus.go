package consensus

import (
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/url"
	"time"
)

func Init(in chan url.Status) (statusReport chan error, err error) {
	statusReport = make(chan error, 10)
	go checkConsensus(in, statusReport)

	return statusReport, nil
}

func checkConsensus(in chan url.Status, statusReport chan error) {
	for {
		var urlStatuses url.Statuses
		interval := time.After(60 * time.Second)
	timer:
		for {
			select {
			case <-interval:
				if heartbeat.Self.Coordinator { // Only the coordinator checks URL statuses
					deduped := urlStatuses.Dedupe()
					_ = deduped.GetFailed() // TODO: call alert on any contacts with a name that matches failed URL tests
				}
				break timer
			case s := <-in:
				urlStatuses = append(urlStatuses, s)
			}
		}
	}
}
