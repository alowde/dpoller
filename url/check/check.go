// Package check defines the core URL-check configuration structures and associated functions. It includes both the
// configuration and result structures.
package check

import (
	"errors"
	"github.com/alowde/dpoller/node"
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

// Check defines the configuration for a single URL to be checked together with its pass/fail conditions and alerting
// information.
type Check struct {
	URL            string   `json:"url"`
	Name           string   `json:"name"`
	OkStatus       []int    `json:"ok-statuses"`
	AlertThreshold int8     `json:"alert-below"`
	AlertInterval  int      `json:"alert-interval"`
	TestInterval   int      `json:"test-interval"`
	Contacts       []string `json:"contacts"`
}

// run runs a single URL test.
func (t Check) run() (s Status) {
	time_start := time.Now()
	resp, err := client.Get(t.URL)
	s = Status{
		Node:      node.Self,
		Url:       t,
		Rtime:     int(time.Since(time_start) / 1000000),
		Timestamp: int(time.Now().Unix()),
	}
	if err != nil {
		s.StatusCode = 0
		s.StatusTxt = err.Error()
	} else {
		defer resp.Body.Close()
		s.StatusCode = resp.StatusCode
		s.StatusTxt = resp.Status
	}
	return
}

// RunAsync runs a single URL test asynchronously and returns a result on the
// provided channel.
func (t Check) RunAsync(c chan Status) {
	go func(chan Status) {
		defer close(c)
		timeStart := time.Now()
		resp, err := client.Get(t.URL)
		s := Status{
			Node:      node.Self,
			Url:       t,
			Rtime:     int(time.Since(timeStart) / 1000000),
			Timestamp: int(time.Now().Unix()),
		}
		if err != nil {
			s.StatusCode = 0
			s.StatusTxt = err.Error()
		} else {
			defer resp.Body.Close()
			s.StatusCode = resp.StatusCode
			s.StatusTxt = resp.Status
		}
		c <- s
	}(c)
}

type Checks []Check

// Run exposes the testing functionality for an array of URL tests, allowing
// them to be conducted simultaneously.
func (t Checks) Run() (s Statuses) {
	testCount := len(t)
	results := make(chan Status, testCount)
	for i, v := range t {
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

func (c *Checks) ByName(name string) (result Check, err error) {
	for _, v := range *c {
		if v.Name == name {
			return v, nil
		}
	}
	return Check{}, errors.New("name not found")
}
