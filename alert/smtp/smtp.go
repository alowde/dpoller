package smtp

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/alert"
	"github.com/alowde/dpoller/node"
	"net/smtp"
)

var config struct {
	Server   string `json:"server"`
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
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Server)
	err := smtp.SendMail(config.Server, auth, "dpoller@example.com", to, msg)
	return err
}

func (c smtpContact) GetName() string {
	return c.Name
}

func initialise(message json.RawMessage, ll logrus.Level) error {
	logrus.SetLevel(ll)

	log = logrus.WithFields(logrus.Fields{
		"routine": "smtpAlert",
		"ID":      node.Self.ID,
	})

	if err := json.Unmarshal(message, &config); err != nil {
		return err
	}
	log.Debug("Successfully received SMTP config")
	return nil
}

func parseContact(message json.RawMessage) (contact alert.Contact, err error) {
	var S smtpContact
	if err := json.Unmarshal(message, &S); err != nil {
		return nil, err
	}
	return S, nil
}

func init() {
	alert.RegisterConfigFunction("smtp", initialise)
	alert.RegisterContactFunction("smtp", parseContact)
}
