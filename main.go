package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/alert"
	"github.com/alowde/dpoller/config"
	"github.com/alowde/dpoller/consensus"
	"github.com/alowde/dpoller/coordinate"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/listen"
	"github.com/alowde/dpoller/node"
	"github.com/alowde/dpoller/publish"
	"github.com/alowde/dpoller/url"
	"github.com/alowde/dpoller/url/urltest"
	"github.com/pkg/errors"
	"time"
)

var routineStatus map[string]chan error
var heartbeatResult chan error

var schan chan urltest.Status // schan passes individual status messages from listeners and test to consensus evaluation
var hchan chan heartbeat.Beat // hchan passes heartbeats from listeners and heartbeat to coordinator

var logger = log.WithField("routine", "main")

func initialise() error {

	// TODO: allow log level change
	logger.Level = log.DebugLevel

	routineStatus = make(map[string]chan error)

	heartbeatResult = make(chan error)

	var err error

	if err := config.Load(); err != nil {
		return errors.Wrap(err, "could not load config")
	}
	if err := node.Initialise(); err != nil {
		return errors.Wrap(err, "could not initialise node data")
	}
	if routineStatus["listen"], hchan, schan, err = listen.Init(*config.Unparsed.Listen); err == nil {
		if routineStatus["coordinate"], err = coordinate.Init(hchan); err != nil {
			return errors.Wrap(err, "could not initialise coordinator routine")
		}
		if routineStatus["consensus"], err = consensus.Init(schan); err != nil {
			return errors.Wrap(err, "could not initialise consensus monitoring routine")
		}
	} else {
		return errors.Wrap(err, "could not initialise listen functions")
	}
	if err := publish.Init(*config.Unparsed.Publish, hchan, schan); err != nil {
		return errors.Wrap(err, "could not initialise publish functions")
	}
	if routineStatus["Url"], err = url.Init(*config.Unparsed.Tests, schan); err != nil {
		return errors.Wrap(err, "could not initialise URL testing functions")
	}
	if err := alert.Init(*config.Unparsed.Alert, *config.Unparsed.Contacts); err != nil {
		return errors.Wrap(err, "could not initialise alert function")
	}
	return nil
}

func main() {
	if err := initialise(); err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	go checkHeartbeats(heartbeatResult, routineStatus)

	/* rubbish to be logged better
	fmt.Println(node.Self)
	fmt.Println(heartbeat.NewBeat())
	fmt.Println(url.Tests)
	*/
	fmt.Printf("End %+v", <-heartbeatResult)

}

// checkHeartbeats
func checkHeartbeats(result chan error, statusChans map[string]chan error) {
	var routineStatus = make(map[string]error)
	for {
		wait := time.After(60 * time.Second)
		<-wait
		// check each routine's status, getting only the most recent status for each
		for k, c := range statusChans {
			for i := 0; i < len(c); i++ {
				routineStatus[k] = <-c
			}
			switch v := routineStatus[k].(type) {
			case heartbeat.RoutineNormal:
				//check if the last normal status is too long ago
				if time.Since(v.Timestamp).Seconds() > 60 {
					result <- errors.Wrapf(v, "Routine %v timed out", k)
				}
			default:
				// some kind of error
				fmt.Printf("died due to unknown error %v from %v\n", v, k)
				result <- errors.Wrapf(v, " From routine: %v", k)
				//close(result)
				return
			}
		}
		if err := publish.Publish(heartbeat.NewBeat(), time.After(10*time.Second)); err != nil {
			fmt.Println("died due to can't publish")
			result <- err
			close(result)
			return
		}
	}
}
