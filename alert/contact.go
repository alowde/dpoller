package alert

import (
	"encoding/json"
	"github.com/alowde/dpoller/url/check"
)

// Contact describes a generic alertable endpoint, and can be extended to include any alert mechanism.
type Contact interface {
	SendAlert(check check.Check, result check.Result) error
	GetName() string
}

var contacts []Contact

type contactParseFunction func(message json.RawMessage) (contact Contact, err error)

// RegisterContactFunction is called as a side-effect of importing an alert mechanism. It accepts a lambda that will be
// supplied with the details of each known contact.
func RegisterContactFunction(name string, f contactParseFunction) {
	contactParseFunctions[name] = f
}

var contactParseFunctions = make(map[string]contactParseFunction)
