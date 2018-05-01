// Package listen provides a generic interface for receiving check results and heartbeat messages from other nodes.
// It's expected that a listener will receive from all other nodes, though it may not connect to every one.
// Listeners are one of four routines that must send a heartbeat for the node to be considered healthy. This allows for
// broker/external connectivity checks to be easily incorporated into the node's self-check process.
package listen

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/logger"
	"github.com/alowde/dpoller/url/check"
	"github.com/pkg/errors"
)

var log *logrus.Entry

type configParseFunction func(message json.RawMessage, ll logrus.Level) (watchdog chan error, hchan chan heartbeat.Beat, schan chan check.Status, err error)

// RegisterConfigFunction is called as a side-effect of importing a listener module. It accepts a lambda that will have
// all related configuration passed to it.
func RegisterConfigFunction(name string, f configParseFunction) {
	configParseFunctions[name] = f
}

var configParseFunctions = make(map[string]configParseFunction)

// Initialise distributes configuration to the imported listener modules by calling their registered config functions.
func Initialise(config json.RawMessage, ll logrus.Level) (watchdog chan error, hchan chan heartbeat.Beat, schan chan check.Status, err error) {

	log = logger.New("listen", ll)

	log.Debug("Parsing listen configuration")

	watchdog = make(chan error)
	hchan = make(chan heartbeat.Beat)
	schan = make(chan check.Status)

	// Unpack JSON only one level to allow plugins to define their own schema
	var C map[string]json.RawMessage
	if err := json.Unmarshal(config, &C); err != nil {
		return nil, nil, nil, errors.Wrap(err, "could not parse listener configuration collection")
	}

	// Iterate over configs received and pass them to registered modules
	var hasOkListener bool
	for listenerName, rawConfig := range C {

		// If we don't have any listener plugins by this name then skip it
		if _, ok := configParseFunctions[listenerName]; !ok {
			log.WithField("config name", listenerName).
				Warn("Found unused listener config")
			continue
		}

		// Call the provided configuration function and link the received channels to our aggregate channel
		log.WithField("name", listenerName).
			Debug("Configuring listener module")
		w, h, s, err := configParseFunctions[listenerName](rawConfig, ll)
		if err != nil {
			log.WithField("name", listenerName).
				WithField("received error", err).
				Warn("Received an error while providing configuration to listener module")
			continue
		}
		go func(in, out chan error) {
			for {
				out <- <-in
			}
		}(w, watchdog)
		go func(in, out chan heartbeat.Beat) {
			for {
				out <- <-in
			}
		}(h, hchan)
		go func(in, out chan check.Status) {
			for {
				out <- <-in
			}
		}(s, schan)
		hasOkListener = true
		delete(configParseFunctions, listenerName)
	}

	// Any configuration functions left haven't been successfully initialised
	for listenerName := range configParseFunctions {
		log.WithField("listener name", listenerName).
			Warn("Listener module found no config")
	}
	if hasOkListener {
		log.Debug("Configured listeners")
		return watchdog, hchan, schan, nil
	}
	return nil, nil, nil, errors.New("No configuration matched listener modules")
}
