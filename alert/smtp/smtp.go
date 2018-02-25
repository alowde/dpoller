package smtp

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/alert"
	"github.com/alowde/dpoller/logger"
	"net/smtp"
)

// Config describes an SMTP relay host, used for sending alerts.
var Config struct {
	Server   string `json:"server"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var log *logrus.Entry

type smtpContact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (c smtpContact) SendAlert() error {
	smsg := fmt.Sprintf("To: %v\r\n"+
		"Subject: Alert from dpoller\r\n\r\n"+
		"An alert occurred",
		c.Email)
	to := []string{c.Email}
	msg := []byte(smsg)
	auth := smtp.PlainAuth("", Config.Username, Config.Password, Config.Server)
	host := Config.Server + ":" + Config.Port
	err := smtp.SendMail(host, auth, "dpoller@example.com", to, msg)
	return err
}

// GetName satisfies half of the alert.Contact interface and exposes the contact name.
func (c smtpContact) GetName() string {
	return c.Name
}

func initialise(message json.RawMessage, ll logrus.Level) error {

	log = logger.New("smtpAlert", ll)

	if err := json.Unmarshal(message, &Config); err != nil {
		return err
	}
	log.Debug("Successfully received SMTP config")
	return nil
}

func parseContact(message json.RawMessage) (contact alert.Contact, err error) {
	var S smtpContact
	if err := json.Unmarshal(message, &S); err != nil { // TODO: sanity check returned config
		return nil, err
	}
	return S, nil
}

func init() {
	alert.RegisterConfigFunction("smtp", initialise)
	alert.RegisterContactFunction("smtp", parseContact)
}
