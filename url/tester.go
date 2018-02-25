package url

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/logger"
	"github.com/alowde/dpoller/publish"
	"github.com/alowde/dpoller/url/check"
	"github.com/pkg/errors"
	"time"
)

var log *logrus.Entry

type checkRun struct {
	check.Check
	lastRan time.Time
	result  chan check.Status
}

func (tr *checkRun) run() {
	tr.lastRan = time.Now()
	tr.result = make(chan check.Status)
	tr.Check.RunAsync(tr.result)
}

var Checks check.Checks

// Initialise configures this module and returns a status channel for monitoring.
func Initialise(config []byte, ll logrus.Level) (routineStatus chan error, err error) {

	log = logger.New("url", ll)

	if err = json.Unmarshal(config, &Checks); err != nil {
		return nil, errors.Wrap(err, "unable to parse URL config")
	}

	if log.Level == logrus.DebugLevel {
		for _, v := range Checks {
			log.WithField("routine", "test").Debug(v)
		}
	}
	routineStatus = make(chan error, 300)
	go runTests(routineStatus)
	return routineStatus, nil
}

func runTests(routineStatus chan error) {
	var runList []checkRun
	// Spread initial test runs over a minute to avoid a thundering herd.
	for i := 0; i < 5; i++ {
		minWait := time.After(12 * time.Second)
		for j := 0 + i; j < len(Checks); j += 5 {
			tr := checkRun{
				Check: check.Check{
					URL:           Checks[j].URL,
					Name:          Checks[j].Name,
					OkStatus:      Checks[j].OkStatus,
					AlertInterval: Checks[j].AlertInterval,
					TestInterval:  Checks[j].TestInterval,
					Contacts:      Checks[j].Contacts,
				},
			}
			tr.run()
			runList = append(runList, tr)
		}
		s := heartbeat.NewRoutineNormal()
		s.SetOrigin("init routine")
		routineStatus <- s
		<-minWait
	}

	for {
		minWait := time.After(15 * time.Second)
		// For each test publish any returned results and re-launch the test if required
		for k, tr := range runList {
			select {
			case result := <-tr.result:
				// TODO: Implement async and batch publish methods and use one of those
				log.WithField("url", tr.URL).Debug("Got a result")
				if err := publish.Publish(result, time.After(10*time.Second)); err != nil {
					// TODO: unwrap error and handle timeouts differently from other errors
					log.WithField("error", err).Warn("failed to publish test result")
				}
				log.WithField("url", tr.URL).Debug("Published a result")
				if time.Now().Sub(tr.lastRan) > (60 * time.Second) {
					runList[k].run()
				}
			default:
				// We trust that the Go http client will return a timeout if appropriate
				// and never freeze or fail to return a result. If this proves optimistic
				// we'll need to add a check here for timed out routines.
				//continue
			}
		}
		s := heartbeat.NewRoutineNormal()
		routineStatus <- s.SetOrigin("main routine")
		<-minWait
	}
}
