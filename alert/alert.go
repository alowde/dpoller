package alert

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/logger"
	"github.com/alowde/dpoller/url/urltest"
	"github.com/pkg/errors"
)

type Contact interface {
	SendAlert() error
	GetName() string
}

type configParseFunction func(message json.RawMessage, ll logrus.Level) error
type contactParseFunction func(message json.RawMessage) (contact Contact, err error)

var configParseFunctions = make(map[string]configParseFunction)
var contactParseFunctions = make(map[string]contactParseFunction)

func RegisterConfigFunction(name string, f configParseFunction) {
	configParseFunctions[name] = f
}
func RegisterContactFunction(name string, f contactParseFunction) {
	contactParseFunctions[name] = f
}

var contacts []Contact
var log *logrus.Entry

func Init(contactJson json.RawMessage, alertJson json.RawMessage, ll logrus.Level) error {

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
	return nil
}

func ProcessAlerts(urls urltest.Statuses) error {
	for _, u := range urls {
		for _, uc := range u.Url.Contacts {
			for _, c := range contacts {
				if uc == c.GetName() {
					c.SendAlert()
				}
			}
		}
	}
	return nil
}
