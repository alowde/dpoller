package smtp

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/node"
	"net/smtp"
)

type smtpContact struct {
	name  string
	email string
}

var Config struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var log *logrus.Entry

func (c smtpContact) SendAlert() error {
	smsg := fmt.Sprintf("To: %v\r\n"+
		"Subject: Alert from dpoller\r\n\r\n"+
		"An alert occurred",
		c.email)
	to := []string{c.email}
	msg := []byte(smsg)
	auth := smtp.PlainAuth("", Config.Username, Config.Password, Config.Server)
	err := smtp.SendMail(Config.Server, auth, "dpoller@example.com", to, msg)
	return err
}

func (c smtpContact) GetName() string {
	return c.name
}

// Initialise sets configuration for the package associated with this contact
func (c smtpContact) Initialise(message json.RawMessage, ll logrus.Level) error {
	logrus.SetLevel(ll)

	log = logrus.WithFields(logrus.Fields{
		"routine": "smtpAlert",
		"ID":      node.Self.ID,
	})

	if err := json.Unmarshal(message, &Config); err != nil {
		return err
	}
	log.Debug("Successfully handled a configuration as an SMTP config")
	return nil
}

// ParseContacts processes an array of JSON objects and returns any that are smtpContacts
func ParseContacts(msg []json.RawMessage) (sca []smtpContact) {
	for _, v := range msg {
		var S smtpContact
		if err := json.Unmarshal(v, &S); err != nil {
			break // continue attempting to parse other JSON objects
		}
		sca = append(sca, S)
	}
	return []smtpContact{}
}
