package alert

import "github.com/alowde/dpoller/url/check"

// Send requests an alert for any configured contacts, passing on check & result information
func Send(c check.Check, r check.Result) error {
	for _, uc := range c.Contacts { // For each contact in the check config
		for _, contact := range contacts { // If we have a matching contact name
			if uc == contact.GetName() {
				if err := contact.SendAlert(c, r); err != nil { // Attempt to send an alert
					log.WithField("error", err).Warn("Couldn't send alert message")
				}
			}
		}
	}
	return nil
}
