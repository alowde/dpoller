package main

import (
	"fmt"
	"github.com/alowde/dpoller/alert"
	"github.com/alowde/dpoller/config"
	"github.com/alowde/dpoller/consensus"
	"github.com/alowde/dpoller/coordinate"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/listen"
	"github.com/alowde/dpoller/node"
	"github.com/alowde/dpoller/publish"
	"github.com/alowde/dpoller/url"
	"github.com/pkg/errors"
	"time"
)

var routineStatus map[string]chan error

var heartbeatRoutineStatus chan error

func initialise() error {

	routineStatus = make(map[string]chan error)
	routineStatus["Url"] = make(chan error, 10)

	heartbeatRoutineStatus = make(chan error)

	var err error
	var hchan chan heartbeat.Beat
	var schan chan url.Status

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
	if err := publish.Init(*config.Unparsed.Publish); err != nil {
		return errors.Wrap(err, "could not initialise publish functions")
	}
	if err := url.Init(*config.Unparsed.Tests); err != nil {
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
	go urlRoutine()
	go heartbeatRoutine(heartbeatRoutineStatus, routineStatus)

	fmt.Println(node.Self)
	fmt.Println(heartbeat.NewBeat())
	fmt.Println(url.Tests)
	fmt.Printf("End %+v", <-heartbeatRoutineStatus)

}

func urlRoutine() {
	for {
		minWait := time.After(60 * time.Second) // TODO: allow individual URLS to specify an interval
		for _, v := range url.RunTests() {
			if err := publish.Publish(v, time.After(10*time.Second)); err != nil {
				// TODO: unwrap error and handle timeouts differently from other errors
				fmt.Printf("%v\n", err)
				return
			}
			routineStatus["Url"] <- heartbeat.RoutineNormal{time.Now()}
		}
		<-minWait
	}
}

func heartbeatRoutine(result chan error, statusChans map[string]chan error) {
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
					fmt.Println("died due to routine timed out")
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
