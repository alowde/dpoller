package amqp

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/logger"
	"github.com/alowde/dpoller/publish"
	"github.com/pkg/errors"
)

// Broker holds the configuration and state of the AMQP broker
var Broker = &broker{}

var log *logrus.Entry

func initialise(config json.RawMessage, ll logrus.Level) error {

	log = logger.New("amqpPublish", ll)

	log.Debug("Initialising publisher")
	if err := json.Unmarshal(config, Broker); err != nil {
		return errors.Wrap(err, "could not parse configuration")
	}
	if err := Broker.validate(); err != nil {
		return errors.Wrap(err, "invalid configuration")
	}
	log.Debug("Connecting to AMQP broker")
	if err := Broker.connect(); err != nil {
		return errors.Wrap(err, "error connecting to AMQP broker")
	}
	return nil
}

func init() {
	publish.RegisterConfigFunction("amqp", initialise)
	publish.RegisterStatusPublishFunction("amqp", sendStatus)
	publish.RegisterHeartbeatPublishFunction("amqp", sendHeartbeat)
}
