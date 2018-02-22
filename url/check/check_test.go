package check

import "testing"
import (
	"github.com/alowde/dpoller/node"
	"net"
	"time"
)

var time1 int
var time2 int

var node1 = node.Node{
	1000000000000000000,
	net.IP{10, 0, 0, 1},
	"test_node_1",
}
var node2 = node.Node{
	2000000000000000000,
	net.IP{10, 0, 0, 2},
	"test_node_2",
}

/*
type Check struct {
	URL           string   `json:"url"`
	Name          string   `json:"name"`
	OkStatus      []int    `json:"ok-statuses"`
	AlertInterval int      `json:"alert-interval"`
	TestInterval  int      `json:"test-interval"`
	Contacts      []string `json:"contacts"`
}
*/

// Valid url, should pass
var check1 = Check{
	URL:           "www.google.com",
	Name:          "Google",
	OkStatus:      []int{200},
	AlertInterval: 30,
	TestInterval:  10,
	Contacts:      []string{"ops1", "ops2"},
}

// Valid url, should fail
var check2 = Check{
	URL:           "www.google.com",
	Name:          "Google",
	OkStatus:      []int{700},
	AlertInterval: 30,
	TestInterval:  10,
}

// Unresolvable url
var check3 = Check{
	URL:           "www.thisshouldneverresolvebutifitdoesohboy.wtfiswrongwithiana",
	Name:          "Google",
	OkStatus:      []int{200},
	AlertInterval: 30,
	TestInterval:  10,
	Contacts:      []string{"ops1", "ops2"},
}

/*
type Status struct {
   Node       node.Node
   Url        Check   // the URL that was tested
   Rtime      int    // number of milliseconds taken to complete the request
   StatusCode int    // status code returned, or magic number 0 for non-numeric status
   StatusTxt  string // detailed description of the status returned
   Timestamp  int    // timestamp at which this status was recorded
}
*/

// A 200 status from node1
var status1 = Status{
	Node:       node1,
	Url:        check1,
	Rtime:      20,
	StatusCode: 200,
	StatusTxt:  "OK",
	Timestamp:  time1,
}

// A 500 status from node1
var status2 = Status{
	Node:       node1,
	Url:        check1,
	Rtime:      20,
	StatusCode: 500,
	StatusTxt:  "status2",
	Timestamp:  time1,
}

func init() {
	t, _ := time.Parse("20060102 150405", "20180101 010000")
	time1 = int(t.Unix())
}

func TestGetFailed(t *testing.T) {

	var f Statuses
	var l int

	testSetOne := Statuses{status1, status2}
	f = testSetOne.GetFailed()
	l = len(f)
	if l != 1 {
		t.Errorf("Error in GetFailed(), expected one failed test, got %v", l)
	} else {
		if f[0].StatusTxt != "status2" {
			t.Errorf("Error in GetFailed(), expected status2, got %#v", f[0])
		}
	}
}
