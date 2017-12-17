package flags

import (
	"flag"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

var MainLog LogLevel
var AlertLog, ConfLog, ConsensusLog, CoordLog, BeatLog, ListenLog, PubLog, UrlLog LogLevel

type LogLevel struct {
	logrus.Level
	set bool
}

func (ll *LogLevel) Set(value string) error {
	var err error
	ll.Level, err = logrus.ParseLevel(value)
	if err != nil {
		return errors.Wrap(err, "while setting value from flag")
	}
	ll.set = true
	return nil
}

func (ll *LogLevel) String() string {
	if ll == nil {
		return "undefined"
	}
	return ll.Level.String()
}

func (ll *LogLevel) Default(value string) error {
	if ll.set {
		return nil
	}
	return ll.Set(value)
}

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

func Fill() {
	if !MainLog.set {
		MainLog.Set("warn")
	}
	for _, v := range [8]*LogLevel{&AlertLog, &ConfLog, &ConsensusLog, &CoordLog, &BeatLog, &ListenLog, &PubLog, &UrlLog} {
		v.Default(MainLog.Level.String())
	}
}
