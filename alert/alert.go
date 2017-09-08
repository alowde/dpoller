package alert

import (
	"encoding/json"
	alertSmtp "github.com/alowde/dpoller/alert/smtp"
	"github.com/pkg/errors"
)

type Contact interface {
	SendAlert() error
	GetName() string
	Initialise(json.RawMessage) error
}

var contacts []Contact

func Init(contactJson json.RawMessage, alertConfig json.RawMessage) error {
	var err error
	if contacts, err = parseContacts(contactJson); err != nil {
		return errors.Wrap(err, "could not parse contact JSON")
	}
	// turn the alert configuration data into an array of individual alert configurations,
	// then attempt to parse each one against the list of contacts (and in turn, alert packages)
	var AlertConfigurations []json.RawMessage
	if err := json.Unmarshal(alertConfig, &AlertConfigurations); err != nil {
		return errors.New("could not parse alert configuration collection (is it an array?)")
	}
	// TODO: implement configs as structs with the specific package type included
	for _, v := range AlertConfigurations {
		for _, c := range contacts {
			if err := c.Initialise(v); err == nil {
				break // a contact handled the configuration, no need to try more
			}
		}
	}
	return nil
}

// parseContacts calls the various packages and returns the abstracted Contact interface set
func parseContacts(message json.RawMessage) (c []Contact, e error) {
	// We can't assign the returned array directly as it doesn't meet the interface requirements, but we can copy individual elements
	s := alertSmtp.ParseContacts(json.RawMessage{})
	for _, v := range s {
		c = append(c, v)
	}
	// Further packages to come...
	return c, nil
}
