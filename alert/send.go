package alert

import (
	"github.com/alowde/dpoller/url/check"
	"time"
)

var notBefore = make(map[string]time.Time)

// Send requests an alert for any configured contacts, passing on check & result information
func Send(c check.Check, r check.Result) error {
	// don't send alerts more often than check.Check.AlertInterval
	if nb, exist := notBefore[c.Name]; !exist || nb.Before(time.Now()) {
		notBefore[c.Name] = time.Now().Add(time.Duration(c.AlertInterval) * time.Second)
		for _, uc := range c.Contacts { // For each contact in the check config
			for _, contact := range contacts { // If we have a matching contact name
				if uc == contact.GetName() {
					if err := contact.SendAlert(c, r); err != nil { // Attempt to send an alert
						log.WithField("error", err).Warn("Couldn't send alert message")
					}
				}
			}
		}
	}
	return nil
}
