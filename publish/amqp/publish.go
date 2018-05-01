package amqp

import (
	"context"
	"encoding/json"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/url/check"
	"github.com/pkg/errors"
)

// sendStatus is a thin wrapper around the Broker, turns the status into a []byte + "status" string
func sendStatus(ctx context.Context, status check.Status) error {

	msg, err := json.Marshal(status)
	if err != nil {
		return errors.Wrap(err, "could not serialise message")
	}

	return Broker.send(ctx, msg, "status")
}

// sendHeartbeat is a thin wrapper around the Broker, turns the status into a []byte + "heartbeat" string
func sendHeartbeat(ctx context.Context, beat heartbeat.Beat) error {

	msg, err := json.Marshal(beat)
	if err != nil {
		return errors.Wrap(err, "could not serialise message")
	}

	return Broker.send(ctx, msg, "heartbeat")
}
