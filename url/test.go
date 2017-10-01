package url

import (
	"encoding/json"
	//	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/node"
	"github.com/pkg/errors"
	"net"
	"net/http"
	"time"
)

var transport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout: 20 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 10 * time.Second,
}
var client = &http.Client{
	Timeout:   time.Second * 60,
	Transport: transport,
}

type Test struct {
	URL           string   `json:"url"`
	Name          string   `json:"name"`
	OkStatus      []int    `json:"ok-statuses"`
	AlertInterval int      `json:"alert-interval"`
	Contacts      []string `json:"contacts"`
}

func (t Test) run() (s Status) {
	time_start := time.Now()
	resp, err := client.Get(t.URL)
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
	for _, v := range Tests {
		log.WithField("routine", "test").Info(v)
	}

	return nil
}

func RunTests() (s Statuses) {
	testCount := len(Tests)
	results := make(chan Status, testCount)
	for i, v := range Tests {
		if i == testCount {
			break // Don't run more tests that we prepared for
		}
		log.Infof("Running test: %v", v)
		go func() { results <- v.run() }() // run tests concurrently
	}
	for i := 0; i < testCount; i++ {
		s = append(s, <-results)
	}
	return s
}
