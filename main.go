package main

import (
	"context"
	"flag"
	"github.com/Sirupsen/logrus"
	_ "github.com/alowde/dpoller/alert/smtp"
	"github.com/alowde/dpoller/config"
	"github.com/alowde/dpoller/heartbeat"
	_ "github.com/alowde/dpoller/listen/amqp"
	"github.com/alowde/dpoller/logger"
	"github.com/alowde/dpoller/node"
	"github.com/alowde/dpoller/pkg/flags"
	"github.com/alowde/dpoller/publish"
	_ "github.com/alowde/dpoller/publish/amqp"
	"time"
)

var log *logrus.Entry

func init() {
	flags.Create()
}

func main() {

	flag.Parse()
	flags.Fill()

	log = logger.New("main", flags.MainLog.Level)

	// Initialise the instance of the application with runtime data - random ID, external IP address etc.
	if err := node.Initialise(flags.ConfLog.Level); err != nil {
		log.Debugf("%+v\n", err)
		log.WithError(err).
			Fatal("Failed to initialise")
	}

	// Load configuration.
	conf, err := config.NewSkeleton(flags.ConfLog.Level)
	if err != nil {
		log.WithError(err).
			Fatal("Failed to load config")
	}

	// Provide received configuration to subroutines and attempt to start them all up.
	r := newRoutines()
	if err := r.start(conf); err != nil {
		log.WithError(err).
			Fatal("could not initialise a subroutine")
	}

	// Now we enter an infinite loop until either a subroutine returns an error, we fail to publish a heartbeat or the
	// application is killed.
	for {
		waitBetweenChecks()

		if err := r.check(); err != nil {
			log.WithError(err).
				Fatal("died due to error from subroutine")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := publish.Publish(ctx, heartbeat.NewBeat()); err != nil {
			cancel()
			log.Fatal("died due to can't publish")
		}
		cancel()
	}
}

// waitBetweenChecks is a basic blocking function that when called will wait the correct amount of time depending on
// whether this node is a coordinator/feasible coordinator or not.
// This might be more idiomatic implemented as a dynamic time.Ticker but this works and is obvious about how it works.
func waitBetweenChecks() {
	var waitTime time.Duration
	if heartbeat.GetCoordinator() || heartbeat.GetFeasibleCoordinator() {
		waitTime = 5 * time.Second
	} else {
		waitTime = 30 * time.Second
	}
	wait := time.After(waitTime)
	<-wait
}
