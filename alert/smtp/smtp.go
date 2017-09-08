package smtp

import (
	"encoding/json"
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
	return nil
}

func (c smtpContact) GetName() string {
	return ""
}

func (c smtpContact) Initialise(message json.RawMessage) error {
	if err := json.Unmarshal(message, &Config); err != nil {
		return err
	}
	return nil
}

func ParseContacts(json json.RawMessage) []smtpContact {
	return []smtpContact{}
}
