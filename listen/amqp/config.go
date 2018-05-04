package amqp

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/listen"
	"github.com/alowde/dpoller/logger"
	"github.com/alowde/dpoller/url/check"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

var log *logrus.Entry

// Config contains all data used to connect to an AMQP broker.
type Config struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Pass     string `json:"pass"`
	Scheme   string `json:"scheme"`
	Port     int    `json:"port"`
	Exchange string `json:"exchange"` // exchange name
	Channel  string `json:"channel"`  // channel name
}

func (c Config) validate() error {
	if len(c.Host) <= 0 {
		return fmt.Errorf("invalid host field")
	}
	if len(c.User) <= 0 {
		return fmt.Errorf("invalid user field")
	}
	if len(c.Pass) <= 0 {
		return fmt.Errorf("invalid pass field")
	}
	return nil
}

// initialise turns the provided config []byte into a validated amqpBroker, generates the listen
// channels and calls listen to spawn a parser for the incoming messages.
func initialise(config json.RawMessage, ll logrus.Level) (result chan error, hchan chan heartbeat.Beat, schan chan check.Status, err error) {

	log = logger.New("amqpListen", ll)

	log.Debug("Initialising AMQP listener")
	brokerInstance, err = newBroker(config)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "could not initialise listener")
	}
	// Connecting here helps detect issues early
	if err := brokerInstance.connect(); err != nil {
		return nil, nil, nil, errors.Wrap(err, "error while connecting listener")
	}

	result = make(chan error, 10)
	hchan = make(chan heartbeat.Beat)
	schan = make(chan check.Status)

	if err := brokerInstance.listen(result, hchan, schan); err != nil {
		return nil, nil, nil, errors.Wrap(err, "error while calling listen function")
	}
	log.Debug("Completed AMQP listener configuration")
	return result, hchan, schan, nil
}

// broker is an active instance of an AMQP broker connection.
type broker struct {
	Config                      // broker configuration
	connection *amqp.Connection // broker connection object
	achannel   *amqp.Channel    // default AMQP channel
	closed     chan *amqp.Error // connection closed flag
}

// newBroker attempts to load and parse a given AMQP config filename and
// returns a resulting Broker object.
func newBroker(config []byte) (*broker, error) {
	var b broker
	var c Config
	if err := json.Unmarshal(config, &c); err != nil {
		return &b, errors.Wrap(err, "unable to parse AMQP config")
	}
	if err := c.validate(); err != nil {
		return &b, errors.Wrap(err, "could not validate config")
	}
	b.Config = c
	b.closed = make(chan *amqp.Error)
	close(b.closed)
	return &b, nil
}

var brokerInstance *broker

func init() {
	listen.RegisterConfigFunction("amqp", initialise)
}
