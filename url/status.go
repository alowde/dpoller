package url

//import "time"
//import "net/http"
import "github.com/alowde/dpoller/node"

type Status struct {
	Node       node.Node
	Url        Test   // the URL that was tested
	Rtime      int    // number of milliseconds taken to complete the request
	StatusCode int    // status code returned, or magic number 0 for non-numeric status
	StatusTxt  string // detailed description of the status returned
	Timestamp  int    // timestamp at which this status was recorded
}

// Statuses is an array of Status
type statuses []Status

// dedupe returns a statuses with only the most recent node-url result tuples
func (s statuses) dedupe() (r statuses) {
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
