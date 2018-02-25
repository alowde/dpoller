package heartbeat

import "testing"
import (
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/node"
	"net"
	"time"
)

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
var node3 = node.Node{
	3000000000000000000,
	net.IP{10, 0, 0, 3},
	"test_node_3",
}
var node4 = node.Node{
	4000000000000000000,
	net.IP{10, 0, 0, 4},
	"test_node_4",
}

var testtime time.Time

var beat1 = Beat{
	node1,
	false,
	false,
	time.Now(),
}

func init() {
	Initialise(logrus.FatalLevel)
	testtime, _ = time.Parse("20060102 150405", "20380119 031408") // bonus test
}

func TestEvaluate(t *testing.T) {

	tables := []struct {
		knownBeats    Beats
		self          Beat
		shouldBeCoord bool
		shouldBeFeas  bool
	}{
		// one node in initial state
		{Beats{Beat{node1, false, false, testtime}}, beat1, false, true},
		// one node after one pass
		{Beats{Beat{node1, false, true, testtime}}, Beat{node1, false, true, testtime}, true, false},
		// two nodes feas, self loser
		{Beats{Beat{node2, false, true, testtime}, Beat{node1, false, true, testtime}}, Beat{node1, false, true, testtime}, false, false},
		// two nodes feas, self winner
		{Beats{Beat{node2, false, true, testtime}, Beat{node1, false, true, testtime}}, Beat{node2, false, true, testtime}, true, false},
		// one coord, one feas, self feasible
		{Beats{Beat{node2, false, true, testtime}, Beat{node1, true, false, testtime}}, Beat{node2, false, true, testtime}, false, true},
	}

	for _, table := range tables {
		Self = table.self
		table.knownBeats.Evaluate()
		if Self.Feasible != table.shouldBeFeas {
			t.Errorf("Error in Evaluate(), Feasible was %t, should be %t", Self.Feasible, table.shouldBeFeas)
		}
	}
}
