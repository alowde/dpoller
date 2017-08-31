package url

import (
	"encoding/json"
	"github.com/alowde/dpoller/node"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

type Test struct {
	URL           string
	Name          string
	Timeout       int64
	OkStatus      []int
	AlertInterval int
	Contacts      []string
}

func (t Test) run() (s Status) {
	time_start := time.Now()
	resp, err := http.Get(t.URL) // TODO: use a custom dialler with timeout, etc defined
	if err == nil {
		defer resp.Body.Close()
		s = Status{
			Node:       node.Self,
			Url:        t,
			Rtime:      int(time.Since(time_start) / 1000000),
			StatusCode: resp.StatusCode,
			StatusTxt:  resp.Status,
			Timestamp:  int(time.Now().Unix()),
		}
	} else {
		s = Status{
			Node:       node.Self,
			Url:        t,
			Rtime:      int(time.Since(time_start) / 1000000),
			StatusCode: 0,
			StatusTxt:  err.Error(),
			Timestamp:  int(time.Now().Unix()),
		}
	}
	return s
}

var Tests []Test

func Init(config []byte) error {
	if err := json.Unmarshal(config, &Tests); err != nil {
		return errors.Wrap(err, "unable to parse URL config")
	}
	return nil
}

// TODO: accept a deadline here and time out if needed
func RunTests() (s Statuses) {
	testCount := len(Tests)
	results := make(chan Status, testCount)
	for i, v := range Tests {
		if i == testCount {
			break // Don't run more tests that we prepared for
		}
		go func() { results <- v.run() }() // run tests concurrently
	}
	for i := 0; i < testCount; i++ {
		s = append(s, <-results)
	}
	return s
}
