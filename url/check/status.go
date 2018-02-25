package check

import (
	"github.com/alowde/dpoller/node"
)

// Status is the result of a single Check.
type Status struct {
	Node       node.Node
	Url        Check  // the URL that was tested
	Rtime      int    // number of milliseconds taken to complete the request
	StatusCode int    // status code returned, or magic number 0 for non-numeric status
	StatusTxt  string // detailed description of the status returned
	Timestamp  int    // timestamp at which this status was recorded
}

// Statuses is an array of Status.
type Statuses []Status

// Result is the aggregation of statuses from one URL, particularly useful for alerting.
type Result struct {
	url             Check // the URL that was checked
	averageResponse int   // average number of milliseconds taken to complete the request
	statusCodes     []int // unique status codes that were seen from this URL
	failed          int   // number of checks that failed
	passed          int   // number of checks that passed
}

// Dedupe returns a Statuses containing only the most recent node-url result tuples.
func (s Statuses) Dedupe() (r Statuses) {
	type tup struct {
		nodeId int64
		url    string
	}
	m := make(map[tup]Status)
	// add values into a map, keeping only the more recent when handling duplicates
	for _, v := range s {
		mytup := tup{v.Node.ID, v.Url.URL}
		if _, ok := m[mytup]; ok {
			if m[mytup].Timestamp < v.Timestamp {
				m[mytup] = v
			}
		} else {
			m[mytup] = v
		}
	}
	// turn the map back into a slice (now missing duplicates) for return
	for _, v := range m {
		r = append(r, v)
	}
	return r
}

// GetFailed returns a Statuses containing only test statuses that didn't receive an Ok response.
func (s Statuses) GetFailed() (r Statuses) {

next:
	for _, v := range s {
		for _, u := range v.Url.OkStatus {
			if v.StatusCode == u {
				continue next
			}
		}
		r = append(r, v)
	}
	return
}
