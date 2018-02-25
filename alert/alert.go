package alert

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/logger"
	"github.com/alowde/dpoller/url/check"
	"github.com/pkg/errors"
)

// Contact describes a generic alertable endpoint, and can be extended to include any alert mechanism.
type Contact interface {
	SendAlert() error
	GetName() string
}

type configParseFunction func(message json.RawMessage, ll logrus.Level) error
type contactParseFunction func(message json.RawMessage) (contact Contact, err error)

var configParseFunctions = make(map[string]configParseFunction)
var contactParseFunctions = make(map[string]contactParseFunction)

// RegisterConfigFunction is called as a side-effect of importing an alert mechanism. It accepts a lambda that will have
// all related configuration passed to it.
func RegisterConfigFunction(name string, f configParseFunction) {
	configParseFunctions[name] = f
}

// RegisterContactFunction is called as a side-effect of importing an alert mechanism. It accepts a lambda that will be
// supplied with the details of each known contact.
func RegisterContactFunction(name string, f contactParseFunction) {
	contactParseFunctions[name] = f
}

var contacts []Contact
var log *logrus.Entry

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

// ProcessAlerts iterates over a slice of Status and sends alerts to each contact.
func ProcessAlerts(urls check.Statuses) error {
	for _, u := range urls {
		for _, uc := range u.Url.Contacts {
			for _, c := range contacts {
				if uc == c.GetName() {
					if err := c.SendAlert(); err != nil {
						log.WithField("error", err).Warn("Couldn't send alert message")
					}
				}
			}
		}
	}
	return nil
}
