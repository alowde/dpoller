package amqp

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/url/check"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"time"
)

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
		return errors.Wrap(err, "could not connect to AMQP broker")
	}

	b.closed = make(chan *amqp.Error)
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

// listen ensures the connection is live and sets up a parsing routine.
func (b *broker) listen(result chan error, hchan chan heartbeat.Beat, schan chan check.Status) error {
	// loop until either the broker connection is not closed or an attempt to open the connection fails
	for {
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
				b.Exchange, // sourceExchange
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
			return nil
		}
	}
}

func parseAmqpMessages(inbox <-chan amqp.Delivery, result chan error, hchan chan heartbeat.Beat, schan chan check.Status) {
	for {
		heartbeatTimer := time.After(15 * time.Second)
	loop:
		for {
			select {
			case <-heartbeatTimer:
				result <- heartbeat.RoutineNormal{Timestamp: time.Now()}
				continue loop
			case message := <-inbox:
				_ = message.Ack(true) // If Ack fails it'll still be easier to deal with elsewhere.
				switch message.Type {
				case "status":
					var s check.Status
					if err := json.Unmarshal(message.Body, &s); err != nil {
						log.WithFields(logrus.Fields{
							"error":    err,
							"delivery": fmt.Sprintf("%#v", message),
						}).Warn("failed to decode a Status delivery, skipping")
						continue
					}
					log.Info("received a Status")
					log.WithFields(logrus.Fields{
						"status": fmt.Sprintf("%#v", s),
					}).Debug("decoded a Status")
					schan <- s
				case "heartbeat":
					var b heartbeat.Beat
					if err := json.Unmarshal(message.Body, &b); err != nil {
						log.WithFields(logrus.Fields{
							"error":    err,
							"delivery": fmt.Sprintf("%#v", message),
						}).Warn("failed to decode a Heartbeat delivery, skipping")
						continue
					}
					log.Info("received a Heartbeat")
					log.WithFields(logrus.Fields{
						"beat": fmt.Sprintf("%#v", b),
					}).Debug("decoded a Heartbeat")
					hchan <- b
				default:
					log.WithFields(logrus.Fields{
						"type": message.Type,
					}).Warn("received unknown delivery type")
				}

			}
		}
	}
}
