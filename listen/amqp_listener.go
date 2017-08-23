package listen

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/url"
	"github.com/pkg/errors"
	samqp "github.com/streadway/amqp"
)

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

// broker contains all of the information required to connect to an AMQP broker
type broker struct {
	Config                       // broker configuration
	connection *samqp.Connection // broker connection object
	achannel   *samqp.Channel    // default AMQP channel
	closed     chan *samqp.Error // connection closed flag
}

// connect establises connection for AMQP broker
func (b *broker) connect() error {
	var err error
	uri := fmt.Sprintf(
		"%v://%v:%v@%v:%v/",
		b.Scheme,
		b.User,
		b.Pass,
		b.Host,
		b.Port,
	)
	if b.connection, err = samqp.Dial(uri); err != nil {
		fmt.Printf("%#v\n", uri)
		return errors.Wrap(err, "could not connect to AMQP broker")
	}

	b.closed = make(chan *samqp.Error)
	b.connection.NotifyClose(b.closed)

	if b.achannel, err = b.connection.Channel(); err != nil {
		return errors.Wrap(err, "could not open AMQP channel")
	}

	// Declare the exchange now just in case it isn't present
	if err = b.achannel.ExchangeDeclare(
		b.Exchange, // name of the exchange
		"topic",    // type is always topic
		true,       // durable
		false,      // delete when complete
		false,      // internal
		false,      // noWait
		nil,        // arguments
	); err != nil {
		return errors.Wrap(err, "could not declare AMQP exchange")
	}
	return nil
}

// listen ensures the connection is live and sets up a parsing routine
func (b *broker) listen(result chan error, hchan chan heartbeat.Beat, schan chan url.Status) error {
	select {
	case <-b.closed:
		if err := b.connect(); err != nil {
			return errors.Wrap(err, "could not connect before listening")
		}
	default:
		// declare a queue on the AMQP broker
		queue, err := b.achannel.QueueDeclare(
			"",    // Ask server to generate a name
			false, // durable
			true,  // delete when unused
			false, // exclusive
			false, // noWait
			nil,   // arguments
		)
		if err != nil {
			return errors.Wrap(err, "unable to declare an AMQP queue")
		}
		// bind that queue to the dpoller exchange
		if err = b.achannel.QueueBind(
			queue.Name, // name of the queue
			"#",        // bindingKey
			"dpoller",  // sourceExchange
			false,      // noWait
			nil,        // arguments
		); err != nil {
			return errors.Wrap(err, "unable to bind to AMQP queue")
		}
		// receive AMQP messages on a new Go channel
		inbox, err := b.achannel.Consume(
			queue.Name, // name
			"",         // auto generated consumerTag,
			false,      // no auto acknowledgements
			true,       // exclusive
			false,      // option not supported
			false,      // receive deliveries immediately
			nil,        // arguments
		)
		if err != nil {
			return errors.Wrap(err, "unable to consume from AMQP queue")
		}
		// launch the actual parsing routine
		go parseAmqpMessages(inbox, result, hchan, schan)
	}
	return nil
}

// newBroker attempts to load and parse a given AMQP config filename and
// returns a resulting Broker object
func newBroker(config []byte) (*broker, error) {
	var raw = []byte(config)
	var b broker
	var c Config
	if err := json.Unmarshal(raw, &c); err != nil {
		return &b, errors.Wrap(err, "unable to parse AMQP config")
	}
	if err := c.validate(); err != nil {
		return &b, errors.Wrap(err, "could not validate config")
	}
	b.Config = c
	b.closed = make(chan *samqp.Error)
	close(b.closed)
	return &b, nil
}

var brokerInstance *broker

// Init turns the provided config []byte into a validated amqpBroker, generates the listen
// channels and calls listen to spawn a parser for the incoming messages.
func Init(config []byte) (result chan error, hchan chan heartbeat.Beat, schan chan url.Status, err error) {
	brokerInstance, err = newBroker(config)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "could not initialise listener")
	}
	// Connecting here helps detect issues early
	if err := brokerInstance.connect(); err != nil {
		return nil, nil, nil, errors.Wrap(err, "error while connecting listener")
	}

	result = make(chan error)
	hchan = make(chan heartbeat.Beat)
	schan = make(chan url.Status)

	if err := brokerInstance.listen(result, hchan, schan); err != nil {
		return nil, nil, nil, errors.Wrap(err, "error while calling listen function")
	}

	return result, hchan, schan, nil
}

// TODO: Implement watchdog
func parseAmqpMessages(inbox <-chan samqp.Delivery, result chan error, hchan chan heartbeat.Beat, schan chan url.Status) {
	defer func() { close(result) }()
	for msg := range inbox {
		msg.Ack(true)
		switch msg.Type {
		case "status":
			var s url.Status
			if err := json.Unmarshal(msg.Body, &s); err != nil {
				log.WithFields(log.Fields{
					"error":    err,
					"delivery": fmt.Sprintf("%#v", msg),
				}).Warn("failed to decode a Status delivery, skipping")
				continue
			}
			log.WithFields(log.Fields{
				"status": fmt.Sprintf("%#v", s),
			}).Debug("decoded a status")
			schan <- s
		case "heartbeat":
			var b heartbeat.Beat
			if err := json.Unmarshal(msg.Body, &b); err != nil {
				log.WithFields(log.Fields{
					"error":    err,
					"delivery": fmt.Sprintf("%#v", msg),
				}).Warn("failed to decode a Heartbeat delivery, skipping")
				continue
			}
			log.WithFields(log.Fields{
				"beat": fmt.Sprintf("%#v", b),
			}).Debug("decoded a Heartbeat delivery")
			hchan <- b
		default:
			log.WithFields(log.Fields{
				"type": msg.Type,
			}).Warn("received unknown delivery type")
		}
	}
}
