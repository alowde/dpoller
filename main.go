package main

import (
	"flag"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/alert"
	"github.com/alowde/dpoller/config"
	"github.com/alowde/dpoller/consensus"
	"github.com/alowde/dpoller/coordinate"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/listen"
	"github.com/alowde/dpoller/node"
	"github.com/alowde/dpoller/pkg/flags"
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

var log *logrus.Entry

func init() {
	flags.Create()
}

func initialise() error {

	flag.Parse()
	flags.Fill()

	logrus.SetLevel(flags.MainLog.Level)
	log = logrus.WithField("routine", "main")

	routineStatus = make(map[string]chan error)

	heartbeatResult = make(chan error)

	var err error

	if err := config.Load(flags.ConfLog.Level); err != nil {
		return errors.Wrap(err, "could not load config")
	}
	if err := node.Initialise(flags.ConfLog.Level); err != nil {
		return errors.Wrap(err, "could not initialise node data")
	}
	heartbeat.Init(flags.BeatLog.Level)
	if routineStatus["listen"], hchan, schan, err = listen.Init(*config.Unparsed.Listen, flags.ListenLog.Level); err == nil {
		if routineStatus["coordinate"], err = coordinate.Init(hchan, flags.CoordLog.Level); err != nil {
			return errors.Wrap(err, "could not initialise coordinator routine")
		}
		if routineStatus["consensus"], err = consensus.Init(schan, flags.ConsensusLog.Level); err != nil {
			return errors.Wrap(err, "could not initialise consensus monitoring routine")
		}
	} else {
		return errors.Wrap(err, "could not initialise listen functions")
	}
	if err := publish.Init(*config.Unparsed.Publish, hchan, schan, flags.PubLog.Level); err != nil {
		return errors.Wrap(err, "could not initialise publish functions")
	}
	if routineStatus["url"], err = url.Init(*config.Unparsed.Tests, flags.UrlLog.Level); err != nil {
		return errors.Wrap(err, "could not initialise URL testing functions")
	}
	if err := alert.Init(*config.Unparsed.Alert, *config.Unparsed.Contacts, flags.AlertLog.Level); err != nil {
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

	fmt.Printf("End! got result %+v", <-heartbeatResult)

}

// checkHeartbeats
func checkHeartbeats(result chan error, statusChans map[string]chan error) {
	for {
		var routineStatus = make(map[string]error)
		wait := time.After(60 * time.Second)
		<-wait
		for k, c := range statusChans {
		out:
			// dequeue all statuses, keeping the first non-normal one or the
			// most recent normal one if all normal
			for {
				select {
				case status := <-c:
					routineStatus[k] = status
					if _, ok := status.(heartbeat.RoutineNormal); !ok {
						break out
					}
				default:
					break out
				}
			}
			switch v := routineStatus[k].(type) {
			case heartbeat.RoutineNormal:
				// check if the last normal status is too long ago
				if time.Since(v.Timestamp).Seconds() > 120 {
					log.Infof("Current time %v, timestamp %v", time.Now(), v.Timestamp)
					result <- errors.Wrapf(v, "Routine %v timed out", k)
					return
				}
			case nil:
				// no status returned from this routine this round
			default:
				// some kind of error
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
