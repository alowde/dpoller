package main

import (
	"github.com/alowde/dpoller/alert"
	"github.com/alowde/dpoller/config"
	"github.com/alowde/dpoller/consensus"
	"github.com/alowde/dpoller/coordinate"
	"github.com/alowde/dpoller/heartbeat"
	"github.com/alowde/dpoller/listen"
	"github.com/alowde/dpoller/pkg/flags"
	"github.com/alowde/dpoller/publish"
	"github.com/alowde/dpoller/url"
	"github.com/alowde/dpoller/url/check"
	"github.com/pkg/errors"
	"time"
)

// routine holds the status of a long running concurrent routine
type routine struct {
	status      chan error
	lastCheckin time.Time
}

func (r *routine) check() error {
	for {
		// Iterate over all buffered messages in channel
		select {
		case statusMessage := <-r.status:
			status, ok := statusMessage.(heartbeat.RoutineNormal)
			// If we receive a statusMessage and it's not "RoutineNormal" it's an error, so we return it
			if !ok {
				return statusMessage
			}
			// Otherwise it's a RoutineNormal so update the last checkin time. Because channels are FIFO we can just
			// update each time.
			r.lastCheckin = status.Timestamp
		default:
			// Once we've received all status messages then check if we've received one in the last 120 seconds. If not
			// raise an error as the routine is considered to have timed out
			if time.Since(r.lastCheckin).Seconds() > 120 {
				log.Infof("Current time %v, timestamp %v", time.Now(), r.lastCheckin)
				return heartbeat.NewTimeout()
			}
			return nil
		}
	}
}

// newRoutine returns a new routine
func newRoutine() *routine {
	return &routine{
		status:      make(chan error, 10),
		lastCheckin: time.Now(),
	}
}

type routines map[string]*routine

func newRoutines() routines {
	r := make(map[string]*routine)
	r["listen"] = newRoutine()
	r["coordinate"] = newRoutine()
	r["consensus"] = newRoutine()
	r["url"] = newRoutine()
	return r

}

func (r routines) start(conf *config.Skeleton) (err error) {

	var hchan chan heartbeat.Beat
	var schan chan check.Status

	r["listen"].status, hchan, schan, err = listen.Initialise(*conf.Listen, flags.ListenLog.Level)
	if err != nil {
		err = errors.Wrap(err, "could not initialise listen functions")
		return
	}

	r["coordinate"].status, err = coordinate.Initialise(hchan, flags.CoordLog.Level)
	if err != nil {
		err = errors.Wrap(err, "could not initialise coordinator routine")
		return
	}

	r["consensus"].status, err = consensus.Initialise(schan, flags.ConsensusLog.Level)
	if err != nil {
		err = errors.Wrap(err, "could not initialise consensus monitoring routine")
		return
	}

	r["url"].status, err = url.Initialise(*conf.Tests, flags.UrlLog.Level)
	if err != nil {
		err = errors.Wrap(err, "could not initialise URL testing functions")
		return
	}

	heartbeat.Initialise(flags.BeatLog.Level)

	err = publish.Initialise(*conf.Publish, hchan, schan, flags.PubLog.Level)
	if err != nil {
		err = errors.Wrap(err, "could not initialise publish functions")
		return
	}

	err = alert.Initialise(*conf.Contacts, *conf.Alert, flags.AlertLog.Level)
	if err != nil {
		err = errors.Wrap(err, "could not initialise alert function")
		return
	}

	return
}

func (r routines) check() error {
	for name, v := range r {
		if err := v.check(); err != nil {
			return errors.Wrapf(err, "from routine %v", name)
		}
	}
	return nil
}
