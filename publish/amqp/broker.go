package amqp

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"time"
)

// broker contains all of the information required to connect to an AMQP broker.
type broker struct {
	connection *amqp.Connection // broker connection object
	achannel   *amqp.Channel    // default AMQP channel
	closed     chan *amqp.Error // connection closed flag
	Host       string           `json:"host"`
	User       string           `json:"user"`
	Pass       string           `json:"pass"`
	Scheme     string           `json:"scheme"`
	Port       int              `json:"port"`
	Exchange   string           `json:"exchange"` // exchange name
	Channel    string           `json:"channel"`  // channel name
}

// connect establishes connection for AMQP broker.
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
		log.WithFields(logrus.Fields{
			"url":   uri,
			"error": err,
		}).Warn("error while dialling AMQP broker")
		return errors.Wrap(err, "while dialling AMQP broker")
	}
	if b.achannel, err = b.connection.Channel(); err != nil {
		return errors.Wrap(err, "could not open AMQP channel")
	}
	// Best practice for AMQP is to unconditionally declare the exchange on connection
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

func (b *broker) validate() error {
	if b.Host == "" {
		return errors.New("missing host field")
	}
	if b.User == "" {
		return errors.New("missing user field")
	}
	if b.Pass == "" {
		return errors.New("missing pass field")
	}
	if b.Scheme == "" {
		b.Scheme = "amqp"
	}
	if b.Port == 0 {
		b.Port = 5672
	}
	return nil
}

func (b *broker) send(ctx context.Context, msg []byte, msgType string) error {

	amqpMsg := amqp.Publishing{
		DeliveryMode: amqp.Transient,
		Timestamp:    time.Now(),
		ContentType:  "text/plain",
		Type:         msgType,
		Body:         msg,
	}

	// Loop until we successfully publish, permanently fail to connect, or run out of time
	// Note! The amqp Publish method is blocking and in theory could cause a goroutine leak
	for {
		select {
		case <-ctx.Done():
			return errors.New("deadline expired while publishing to amqp")
		case <-b.closed:
			if err := b.connect(); err != nil {
				return errors.Wrap(err, "could not connect before publishing to amqp")
			}
		default:
			err := b.achannel.Publish(b.Exchange, "status", false, false, amqpMsg)
			if err == nil {
				return nil
			}
			log.WithError(err).
				WithField("message", msg).
				Warn("received error while publishing message")
			// close the connection so that we can try again cleanly
			if err := b.connection.Close(); err != nil {
				log.Warn("AMQP connection may not have closed cleanly")
			}
		}
	}
}
