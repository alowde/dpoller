package publish

import (
	"encoding/json"
	"fmt"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/url"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"time"
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
	Config                      // broker configuration
	connection *amqp.Connection // broker connection object
	achannel   *amqp.Channel    // default AMQP channel
	closed     chan *amqp.Error // connection closed flag
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
	if b.connection, err = amqp.Dial(uri); err != nil {
		fmt.Printf("%#v\n", uri)
		return errors.Wrap(err, "could not connect to AMQP broker")
	}

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
	b.closed = make(chan *amqp.Error)
	b.achannel.NotifyClose(b.closed)

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
	b.closed = make(chan *amqp.Error)
	close(b.closed)
	return &b, nil
}

var brokerInstance *broker

// Init turns the provided config []byte into a validated amqpBroker and connects
func Init(config []byte) (err error) {
	brokerInstance, err = newBroker(config)
	if err != nil {
		return errors.Wrap(err, "could not initialise publisher")
	}
	// Connecting here helps detect issues early
	if err := brokerInstance.connect(); err != nil {
		return errors.Wrap(err, "error while connecting publisher")
	}
	return nil
}

// listen ensures the connection is live and sets up a parsing routine
//func (b *broker) publishHeartbeat(beat heartbeat.Beat, deadline <-chan time.Time) error {
func Publish(i interface{}, deadline <-chan time.Time) error {
	var msgtype string

	switch i.(type) {
	case heartbeat.Beat:
		msgtype = "heartbeat"
	case url.Status:
		msgtype = "status"
	default:
		return errors.New("unknown type of message")
	}

	msgbody, err := json.Marshal(i)
	if err != nil {
		return errors.Wrap(err, "could not serialise message")
	}

	msg := amqp.Publishing{
		DeliveryMode: amqp.Transient,
		Timestamp:    time.Now(),
		ContentType:  "text/plain",
		Type:         msgtype,
		Body:         []byte(msgbody),
	}

	for { // loop until we successfully publish, permanently fail to connect, or run out of time
		select {
		case <-deadline:
			return errors.New("deadline expired while publishing to amqp")
		case <-brokerInstance.closed:
			if err := brokerInstance.connect(); err != nil {
				return errors.Wrap(err, "could not connect before publishing to amqp")
			}
		default:
			if err := brokerInstance.achannel.Publish(brokerInstance.Config.Exchange, msgtype, false, false, msg); err == nil {
				return nil
			} else {
				return errors.Wrap(err, "failed to publish to amqp")
			}
		}
	}
	return nil
}
