package alert

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	alertSmtp "github.com/alowde/dpoller/alert/smtp"
	"github.com/alowde/dpoller/node"
	"github.com/alowde/dpoller/url/urltest"
	"github.com/mattn/go-colorable"
	"github.com/pkg/errors"
)

type Contact interface {
	SendAlert() error
	GetName() string
	Initialise(json.RawMessage, logrus.Level) error
}

var contacts []Contact
var log *logrus.Entry

func Init(contactJson json.RawMessage, alertConfig json.RawMessage, ll logrus.Level) error {

	var logger = logrus.New()
	logger.Formatter = &logrus.TextFormatter{ForceColors: true}
	logger.Out = colorable.NewColorableStdout()
	logger.SetLevel(ll)

	log = logger.WithFields(logrus.Fields{
		"routine": "alert",
		"ID":      node.Self.ID,
	})

	var err error

	// process contact configuration
	var Cja []json.RawMessage
	if err := json.Unmarshal(contactJson, &Cja); err != nil {
		return errors.Wrap(err, "could not parse contact configuration collection (is it an array?)")
	}
	if contacts, err = parseContacts(Cja); err != nil {
		return errors.Wrap(err, "could not parse contact JSON")
	}
	// turn the alert configuration data into an array of individual alert configurations,
	// then attempt to parse each one against the list of contacts (and in turn, alert packages)
	var AlertConfigurations []json.RawMessage
	if err := json.Unmarshal(alertConfig, &AlertConfigurations); err != nil {
		return errors.New("could not parse alert configuration collection (is it an array?)")
	}
	// TODO: implement configs as structs with the specific package type included
	log.Debug("Attempting to configure alert packages")

handled:
	for _, v := range AlertConfigurations {
		log.WithField("config", string(v)).Debug("Processing configuration")
		for _, c := range contacts {
			if err := c.Initialise(v, ll); err == nil {
				break handled // a contact handled the configuration, no need to try more
			}
		}
		log.WithField("config", string(v)).Warn("No alert handler accepted config")
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

// parseContacts calls the various packages and returns the abstracted Contact interface set
func parseContacts(message []json.RawMessage) (c []Contact, e error) {
	log.Debug(string(message[0]))
	// We can't assign the returned array directly as it doesn't meet the interface requirements, but we can copy individual elements
	s := alertSmtp.ParseContacts(message)
	log.WithField("successfully parsed", len(s)).Debug("parsed SMTP contacts")
	for _, v := range s {
		c = append(c, v)
	}
	log.WithField("successfully parsed", len(c)).Info("parsed all contacts")
	// Further packages to come...
	return c, nil
}
