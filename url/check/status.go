package check

import (
	"errors"
	"github.com/alowde/dpoller/node"
	"net"
	"sort"
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

func (s *Status) failed() bool {
	for _, u := range s.Url.OkStatus {
		if s.StatusCode == u {
			return false
		}
	}
	return true
}

// Statuses is an array of Status.
type Statuses []Status

// Result is an aggregation of statuses, particularly useful for alerting.
type Result struct {
	AverageResponse int   // average number of milliseconds taken to complete the request
	StatusCodes     []int // unique status codes that were seen from this URL
	Failed          int   // number of checks that Failed
	Passed          int   // number of checks that Passed
	Total           int   // total number of checks
	PassPercent     int8  // pass percentage, rounded up to whole number
	FailNodeIPs     []net.IP
	FailNodeNames   []string
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
		mytup := tup{v.Node.ID, v.Url.Name}
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

// PerCheckName groups statuses by the check name
func (s *Statuses) PerCheckName() (r map[string]Statuses) {
	r = make(map[string]Statuses)
	for _, v := range *s {
		r[v.Url.Name] = append(r[v.Url.Name], v)
	}
	return r
}

// CalculateResult produces a single Result from Statuses
func (s *Statuses) CalculateResult() (r Result, e error) {
	if len(*s) < 1 {
		return Result{}, errors.New("empty statuses array")
	}
	var failed Statuses
	r.StatusCodes = make([]int, len(*s))
	for i, v := range *s {
		r.AverageResponse = r.AverageResponse + v.Rtime
		r.StatusCodes[i] = v.StatusCode
		if v.failed() {
			failed = append(failed, v)
			r.FailNodeIPs = append(r.FailNodeIPs, v.Node.EIP)
			r.FailNodeNames = append(r.FailNodeNames, v.Node.Name)
		}
	}
	r.AverageResponse = r.AverageResponse / len(*s)
	r.Total = len(*s)
	r.Failed = len(failed)
	r.Passed = r.Total - r.Failed
	r.StatusCodes = uniqInt(r.StatusCodes)
	r.PassPercent = int8((float64(r.Passed) / float64(r.Total)) * float64(100))

	return r, nil
}

// just for fun. Pays the sort price of O(n*log(n)) calls to swap and less, then one allocation per duplicate entry
// I'm assuming this is cheaper than just allocating once for each non-duplicate entry. Need to benchmark
func uniqInt(in sort.IntSlice) (out []int) {
	in.Sort()
	for i := 0; i+1 < len(in); i++ {
		if in[i] == in[i+1] {
			in = append(in[:i], in[i+1:]...)
			i -= 1
		}
	}
	return in
}
