package alert

import "fmt"

type Contact interface {
	SendAlert() error
	GetName() string
}

// GetContacts calls the initialisation function for each of the available contact types.
// If a new contact type is added then the initialisation function will need to be added
// here. This was done manually to avoid the perceived messiness of reflection.
func GetContacts() (c []Contact, e error) {
	//	s, err := getSmsContacts()
	//	if err != nil {
	//		return nil, fmt.Errorf("Unable to load SMS contacts. Received error: %v", err)
	//	}
	//	c = append(c, s)
	s, err := getSmtpContacts()
	if err != nil {
		return nil, fmt.Errorf("Unable to load SMTP contacts. Received error: %v", err)
	}
	for _, v := range s {
		c = append(c, v)
	}
	return c, nil
}
