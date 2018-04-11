// Package alert provides a generic interface for sending alerts to contacts. It allows us to easily write new alert
// methods without modifying the rest of the program.
package alert

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/logger"
	"github.com/pkg/errors"
)

var log *logrus.Entry

type configParseFunction func(message json.RawMessage, ll logrus.Level) error

// RegisterConfigFunction is called as a side-effect of importing an alert mechanism. It accepts a lambda that will have
// all related configuration passed to it.
func RegisterConfigFunction(name string, f configParseFunction) {
	configParseFunctions[name] = f
}

var configParseFunctions = make(map[string]configParseFunction)

// Initialise parses the provided routine configuration.
func Initialise(contactJson json.RawMessage, alertJson json.RawMessage, ll logrus.Level) error {

	log = logger.New("alert", ll)

	log.Debug("Parsing alert configurations")
	var A map[string]json.RawMessage
	if err := json.Unmarshal(alertJson, &A); err != nil {
		return errors.Wrap(err, "could not parse alert configuration collection (is it an array?)")
	}
	for k, m := range A {
		log.WithField("package", k).Debug("Configuring alert package")
		if err := configParseFunctions[k](m, ll); err != nil {
			return errors.Wrap(err, "while processing alert function config")
		}
	}

	log.Debug("Parsing contacts")
	var C map[string][]json.RawMessage
	if err := json.Unmarshal(contactJson, &C); err != nil {
		return errors.Wrap(err, "could not parse contact configuration collection (is it an array?)")
	}
	for k, v := range C {
		for _, c := range v {
			if f, ok := contactParseFunctions[k]; ok {
				contact, err := f(c)
				if err != nil {
					log.Warn("error while trying to process a contact object, ignoring")
					continue
				}
				contacts = append(contacts, contact)
			}
		}
	}
	log.WithField("contact count", len(contacts)).Info("Finished parsing contacts")
	return nil
}
