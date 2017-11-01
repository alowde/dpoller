package url

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/publish"
	"github.com/alowde/dpoller/url/urltest"
	"github.com/pkg/errors"
	"time"
)

var Tests urltest.Tests

func Init(config []byte, result chan urltest.Status) (routineStatus chan error, err error) {
	if err = json.Unmarshal(config, &Tests); err != nil {
		return nil, errors.Wrap(err, "unable to parse URL config")
	}
	for _, v := range Tests { // TODO: rewrite this one as a more useful log
		log.WithField("routine", "test").Info(v)
	}
	routineStatus = make(chan error, 10)
	go runTests(result, routineStatus)
	return routineStatus, nil
}

func runTests(result chan urltest.Status, routineStatus chan error) {
	for {
		minWait := time.After(50 * time.Second) // TODO: allow individual URLS to specify an interval
		//logger.Info("starting tests")
		for _, v := range Tests.Run() {
			//log.WithFields(log.Fields{"test_number": i, "result": v.Url.URL}).Debug("running test")

			if err := publish.Publish(v, time.After(10*time.Second)); err != nil {
				// TODO: unwrap error and handle timeouts differently from other errors
				log.WithField("error", err).Warn("failed to publish test result")
				return
			}

			//logger.Info("finished round of URL tests")
			r := heartbeat.RoutineNormal{Timestamp: time.Now()}
			//r.SetOrigin(v.Url.URL)
			routineStatus <- r
			<-minWait
		}
	}
}
