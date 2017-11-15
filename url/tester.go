package url

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/publish"
	"github.com/alowde/dpoller/url/urltest"
	"github.com/pkg/errors"
	"time"
)

type testRun struct {
	urltest.Test
	lastRan time.Time
	result  chan urltest.Status
}

func (tr *testRun) run() {
	log.Infoln("Run called!")
	//fmt.Println("Run called")
	tr.lastRan = time.Now()
	tr.result = make(chan urltest.Status)
	tr.Test.RunAsync(tr.result)
}

var Tests urltest.Tests

func Init(config []byte) (routineStatus chan error, err error) {
	if err = json.Unmarshal(config, &Tests); err != nil {
		return nil, errors.Wrap(err, "unable to parse URL config")
	}
	for _, v := range Tests { // TODO: rewrite this one as a more useful log
		log.WithField("routine", "test").Info(v)
	}
	routineStatus = make(chan error, 300)
	go runTests(routineStatus)
	return routineStatus, nil
}

func runTests(routineStatus chan error) {
	var runList []testRun
	// Spread initial test runs over a minute to avoid a thundering herd.
	for i := 0; i < 5; i++ {
		minWait := time.After(12 * time.Second)
		for j := 0 + i; j < len(Tests); j += 5 {
			tr := testRun{
				Test: urltest.Test{
					URL:           Tests[j].URL,
					Name:          Tests[j].Name,
					OkStatus:      Tests[j].OkStatus,
					AlertInterval: Tests[j].AlertInterval,
					TestInterval:  Tests[j].TestInterval,
					Contacts:      Tests[j].Contacts,
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

	fmt.Printf("\n\n---\n\n")

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
					fmt.Printf("%v\n", tr.lastRan)
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
		fmt.Println("sent routine with timestamp %v", s.Timestamp)
		<-minWait
	}
}
