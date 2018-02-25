package flags

import (
	"flag"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

var MainLog LogLevel
var AlertLog, ConfLog, ConsensusLog, CoordLog, BeatLog, ListenLog, PubLog, UrlLog LogLevel

// LogLevel is an abstraction of logrus.Level that can be configured with the flags package
type LogLevel struct {
	logrus.Level
	set bool
}

// Set parses a flag-provided value to a logrus.Level
func (ll *LogLevel) Set(value string) error {
	var err error
	ll.Level, err = logrus.ParseLevel(value)
	if err != nil {
		return errors.Wrap(err, "while setting value from flag")
	}
	ll.set = true
	return nil
}

// String safely returns a string description of the Level
func (ll *LogLevel) String() string {
	if ll == nil {
		return "undefined"
	}
	return ll.Level.String()
}

// Default sets a LogLevel level only if it hasn't already been set
func (ll *LogLevel) Default(value string) error {
	if ll.set {
		return nil
	}
	return ll.Set(value)
}

// Create configures the defined flags
func Create() {
	flag.Var(&MainLog, "mainLogLevel", "log level for main routine (debug/info/warn/fatal)")
	flag.Var(&AlertLog, "alertLogLevel", "log level for alert routine (debug/info/warn/fatal)")
	flag.Var(&ConfLog, "confLogLevel", "log level for config routine (debug/info/warn/fatal)")
	flag.Var(&ConsensusLog, "consensusLogLevel", "log level for consensus routine (debug/info/warn/fatal)")
	flag.Var(&CoordLog, "coordinatorLogLevel", "log level for coordinator routine (debug/info/warn/fatal)")
	flag.Var(&BeatLog, "heartbeatLogLevel", "log level for heartbeat routine (debug/info/warn/fatal)")
	flag.Var(&ListenLog, "listenLogLevel", "log level for listen routine (debug/info/warn/fatal)")
	flag.Var(&PubLog, "publishLogLevel", "log level for publish routine (debug/info/warn/fatal)")
	flag.Var(&UrlLog, "urlLogLevel", "log level for url routine (debug/info/warn/fatal)")
}

// Fill initialises the defined flags, defaulting to the level of the Main routine
func Fill() {
	if !MainLog.set {
		MainLog.Set("warn")
	}
	for _, v := range [8]*LogLevel{&AlertLog, &ConfLog, &ConsensusLog, &CoordLog, &BeatLog, &ListenLog, &PubLog, &UrlLog} {
		v.Default(MainLog.Level.String())
	}
}
