package smtp

import (
	"encoding/json"
	"fmt"
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

func (c smtpContact) SendAlert() error {
	smsg := fmt.Sprintf("To: %v\r\n"+
		"Subject: Alert from dpoller\r\n\r\n"+
		"An alert occurred",
		c.email)
	to := []string{c.email}
	msg := []byte(smsg)
	auth := smtp.PlainAuth("", Config.Username, Config.Password, Config.Server)
	err := smtp.SendMail(Config.Server, auth, "dpoller@catapult-elearning.com", to, msg)
	return err
}

func (c smtpContact) GetName() string {
	return ""
}

// Initialise sets configuration for the package associated with this contact
func (c smtpContact) Initialise(message json.RawMessage) error {
	if err := json.Unmarshal(message, &Config); err != nil {
		return err
	}
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
