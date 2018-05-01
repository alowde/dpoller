// Package publish provides a generic interface for sending check result and heartbeat messages to other nodes.
// It's expected that a publisher will publish to all other nodes, though it may not connect to every one.
package publish

import (
	"context"
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/logger"
	"github.com/alowde/dpoller/url/check"
	"github.com/pkg/errors"
	"sync"
	"time"
)

var schan chan check.Status   // channel for internally publishing url statuses
var hchan chan heartbeat.Beat // channel for internally publishing heartbeats
var log *logrus.Entry

type configParseFunction func(message json.RawMessage, ll logrus.Level) error
type statusPublishFunction func(ctx context.Context, status check.Status) error
type heartbeatPublishFunction func(ctx context.Context, beat heartbeat.Beat) error

var configParseFunctions = make(map[string]configParseFunction)
var statusPublishFunctions = make(map[string]statusPublishFunction)
var heartbeatPublishFunctions = make(map[string]heartbeatPublishFunction)

// RegisterConfigFunction is called as a side-effect of importing a publisher module. It accepts a lambda that will have
// all related configuration passed to it, as well as channels for publishing internal messages.
func RegisterConfigFunction(name string, f configParseFunction) {
	configParseFunctions[name] = f
}

func RegisterStatusPublishFunction(name string, f statusPublishFunction) {
	statusPublishFunctions[name] = f
}
func RegisterHeartbeatPublishFunction(name string, f heartbeatPublishFunction) {
	heartbeatPublishFunctions[name] = f
}

func Initialise(config json.RawMessage, hc chan heartbeat.Beat, sc chan check.Status, ll logrus.Level) error {

	hchan = hc
	schan = sc
	log = logger.New("publish", ll)

	log.Debug("Parsing listen configuration")

	// Unpack JSON only one level to allow plugins to define their own schema
	var C map[string]json.RawMessage
	if err := json.Unmarshal(config, &C); err != nil {
		return errors.Wrap(err, "could not parse publisher configuration collection")
	}

	// Iterate over configs received and pass them to registered modules
	var hasOkPublisher bool
	for publisherName, rawConfig := range C {

		// If we don't have any publisher plugins by this name then skip it
		if _, ok := configParseFunctions[publisherName]; !ok {
			log.WithField("config name", publisherName).
				Warn("Found unused publisher config")
			continue
		}

		// Call the provided configuration
		log.WithField("name", publisherName).
			Debug("Configuring publisher module")
		err := configParseFunctions[publisherName](rawConfig, ll)
		if err != nil {
			log.WithField("name", publisherName).
				WithField("received error", err).
				Warn("Received an error while providing configuration to publisher module")
			continue
		}
		hasOkPublisher = true
		delete(configParseFunctions, publisherName)
	}

	// Any configuration functions left haven't been successfully initialised
	for publisherName := range configParseFunctions {
		log.WithField("publisher name", publisherName).
			Warn("Publisher module found no config")
	}

	if hasOkPublisher {
		log.Debug("Configured publishers")
		return nil
	}

	return errors.New("No configuration matched known publisher modules")
}

func Publish(ctx context.Context, i interface{}) error {

	switch v := i.(type) {
	case check.Status:
		log.Debug("publishing a status")
		schan <- v
		return distributeStatuses(ctx, v)

	case heartbeat.Beat:
		log.Debug("publishing a heartbeat")
		hchan <- v
		return distributeHeartbeats(ctx, v)
	default:
		log.WithFields(logrus.Fields{
			"message": i,
		}).Warn("can't publish unknown message type")
		return errors.New("unknown type of message")
	}
}

func distributeHeartbeats(ctx context.Context, beat heartbeat.Beat) error {

	var aggResult = make(chan error, len(heartbeatPublishFunctions))
	var wg sync.WaitGroup

	// Reserve 250ms so publish modules return before the deadline
	deadline, ok := ctx.Deadline()
	if !ok {
		return errors.New("Invalid context provided to publish function")
	}
	timeLeft := deadline.Sub(time.Now().Add(250 * time.Millisecond))
	childCtx, cancel := context.WithTimeout(ctx, timeLeft)
	defer cancel()

	// Call each publish function in parallel and return the results to a buffered channel.
	for _, f := range heartbeatPublishFunctions {
		currentF := f
		wg.Add(1)
		go func(ctx context.Context, beat heartbeat.Beat) {
			aggResult <- currentF(ctx, beat)
			wg.Done()
		}(childCtx, beat)
	}

	// Wait until either context expires or all routines return
	routinesDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(routinesDone)
	}()
	select {
	case <-routinesDone:
	case <-ctx.Done():
		if len(heartbeatPublishFunctions) == 0 {
			return errors.New("No publish functions succeeded before deadline expired")
		}
		// Some publish functions timed out. Should this be an error? Currently is not.
		log.Warn("Not all publish functions responded before deadline expired")
	}

	// Declare a slice with the underlying array the size of our total number of results. This lets us count the number
	// of non-nil responses easily
	var es = make([]error, 0, len(heartbeatPublishFunctions))
	for i := 0; i < len(heartbeatPublishFunctions); i++ {
		if e := <-aggResult; e != nil {
			log.WithError(e).Warn("Received publish function error")
			es = append(es, e)
		}
	}
	// Some functions failed. Should this be an error? Currently is.
	if len(es) > 0 {
		return errors.New("Some publish functions failed")
	}
	return nil
}

func distributeStatuses(ctx context.Context, status check.Status) error {

	var aggResult = make(chan error, len(statusPublishFunctions))
	var wg sync.WaitGroup

	// Reserve 250ms so publish modules return before the deadline
	deadline, ok := ctx.Deadline()
	if !ok {
		return errors.New("Invalid context provided to publish function")
	}
	timeLeft := deadline.Sub(time.Now().Add(250 * time.Millisecond))
	childCtx, cancel := context.WithTimeout(ctx, timeLeft)
	defer cancel()

	// Call each publish function in parallel and return the results to a buffered channel.
	for _, f := range statusPublishFunctions {
		currentF := f
		wg.Add(1)
		go func(ctx context.Context, status check.Status) {
			// Reserve 250ms so this function can aggregate results and return before the deadline
			aggResult <- currentF(ctx, status)
			wg.Done()
		}(childCtx, status)
	}

	// Wait until either context expires or all routines return
	routinesDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(routinesDone)
	}()
	select {
	case <-routinesDone:
	case <-ctx.Done():
		if len(heartbeatPublishFunctions) == 0 {
			return errors.New("No publish functions succeeded before deadline expired")
		}
		// Some publish functions timed out. Should this be an error? Currently is not.
		log.Warn("Not all publish functions responded before deadline expired")
	}

	// Declare a slice with the underlying array the size of our total number of results. This lets us count the number
	// of non-nil responses easily
	var es = make([]error, 0, len(heartbeatPublishFunctions))
	for i := 0; i < len(heartbeatPublishFunctions); i++ {
		if e := <-aggResult; e != nil {
			log.WithError(e).Warn("Received publish function error")
			es = append(es, e)
		}
	}
	// Some functions failed. Should this be an error? Currently is.
	if len(es) > 0 {
		return errors.New("Some publish functions failed")
	}
	return nil
}
