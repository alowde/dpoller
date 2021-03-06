package heartbeat

import (
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/logger"
)

var log *logrus.Entry

// Initialise configures the logging level for the heartbeat module.
func Initialise(ll logrus.Level) {
	log = logger.New("heartbeat", ll)
}
