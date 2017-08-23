package main

import (
	"fmt"
	"github.com/alowde/dpoller/config"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/listen"
	"github.com/alowde/dpoller/node"
	"github.com/alowde/dpoller/publish"
	"github.com/alowde/dpoller/url"
	"github.com/pkg/errors"
	"sourcegraph.com/sqs/goreturns/returns"
	"time"
)

var routineStatusChannels map[string]chan error

var HeartbeatStatus chan error

func initialise() error {

	routineStatusChannels = make(map[string]chan error)
	routineStatusChannels["Listen"] = make(chan error)
	routineStatusChannels["Publish"] = make(chan error)
	routineStatusChannels["Coord"] = make(chan error)
	routineStatusChannels["Url"] = make(chan error)

	HeartbeatStatus = make(chan error)

	if err := config.Load(); err != nil {
		return errors.Wrap(err, "could not load config")
	}
	if err := node.Initialise(); err != nil {
		return errors.Wrap(err, "could not initialise node data")
	}
	_, _, _, err := listen.Init(*config.Unparsed.Listen)
	if err != nil {
		return errors.Wrap(err, "could not initialise listen functions")
	}
	if err := url.Init(*config.Unparsed.Tests); err != nil {
		return errors.Wrap(err, "could not initialise test URL data")
	}
	if err := publish.Init(*config.Unparsed.Publish); err != nil {
		return errors.Wrap(err, "could not initialise publish functions")
	}
	return nil
}

func main() {
	if err := initialise(); err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	go urlRoutine()
	go heartbeatRoutine(HeartbeatStatus, routineStatusChannels)

	fmt.Println(node.Self)
	fmt.Println(heartbeat.NewBeat())
	fmt.Println(url.Tests)
	<-HeartbeatStatus

}

func consensusRoutine() {

}

func coordinatorRoutine() {

	// collect heartbeats for x seconds
	// evaluate our position
	// set global/package state

}

func urlRoutine() {
	for {
		minWait := time.After(60 * time.Second) // TODO: allow individual URLS to specify an interval
		for _, v := range url.RunTests() {
			if err := publish.Publish(v, time.After(10*time.Second)); err != nil {
				// TODO: unwrap error and handle timeouts differently from other errors
				fmt.Printf("%+v\n", err)
			}
			routineStatusChannels["Url"] <- heartbeat.RoutineNormal{time.Now()}
		}
		<-minWait
	}
}

func heartbeatRoutine(result chan error, statusChans map[string]chan error) {
	var routineStatus map[string]error
	for {
		// check each routine's status, getting only the most recent status for each
		for k, c := range statusChans {
			for i := 0; i < len(c); i++ {
				routineStatus[k] = <-c
			}
			switch v := routineStatus[k].(type) {
			case heartbeat.RoutineNormal:
				//check if the last normal status is too long ago
				if time.Since(v.Timestamp).Seconds() > 60 {
					result <- errors.New("Routine timed out")
				}
			default:
				// some kind of error
				result <- v
				close(result)
				return
			}
		}
		if err := publish.Publish(heartbeat.NewBeat(), time.After(10*time.Second)); err != nil {
			result <- err
			close(result)
			return
		}
	}
}
