package heartbeat

import (
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/node"
	"net"
	"testing"
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

/*
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
*/
// All tests use the same simulated time for each heartbeat as time is not currently a factor in tested functions
var testtime, _ = time.Parse("20060102 150405", "20380119 031408") // bonus test

func TestEvaluate(t *testing.T) {

	Initialise(logrus.FatalLevel)

	// Using descriptive variables for the various possible states of the nodes makes the tests clearer
	nodeOneCoordinator := Beat{node1, true, false, testtime}
	nodeOneFeasible := Beat{node1, false, true, testtime}
	nodeOneBoth := Beat{node1, true, true, testtime} // Shouldn't occur in normal operation
	nodeOneNone := Beat{node1, false, false, testtime}
	nodeTwoCoordinator := Beat{node2, true, false, testtime}
	nodeTwoFeasible := Beat{node2, false, true, testtime}
	//	nodeTwoBoth := Beat{node2, true, true, testtime} // Shouldn't occur in normal operation
	nodeTwoNone := Beat{node2, false, false, testtime}
	//	nodeThreeCoordinator := Beat{node3, true, false, testtime}
	//	nodeThreeFeasible := Beat{node3, false, true, testtime}
	//	nodeThreeBoth := Beat{node3, true, true, testtime} // Shouldn't occur in normal operation
	//	nodeThreeNone := Beat{node3, false, false, testtime}
	//	nodeFourCoordinator := Beat{node4, true, false, testtime}
	//	nodeFourFeasible := Beat{node4, false, true, testtime}
	//	nodeFourBoth := Beat{node4, true, true, testtime} // Shouldn't occur in normal operation
	//	nodeFourNone := Beat{node4, false, false, testtime}

	tables := []struct {
		description   string
		knownBeats    Beats
		self          Beat
		shouldBeCoord bool
		shouldBeFeas  bool
	}{
		// Single node tests
		{"one node in initial state", Beats{nodeOneNone}, nodeOneNone, false, true},
		{"one node after one pass", Beats{nodeOneFeasible}, nodeOneFeasible, true, false},
		{"one node after two passes", Beats{nodeOneCoordinator}, nodeOneCoordinator, true, false},
		{"one node is both Feas and Coord", Beats{nodeOneBoth}, nodeOneBoth, true, false},

		// Two node tests
		{"two blank nodes, winners perspective", Beats{nodeOneNone, nodeTwoNone}, nodeOneNone, false, true},
		{"two blank nodes, losers perspective", Beats{nodeOneNone, nodeTwoNone}, nodeTwoNone, false, false},
		{"two feasible nodes pass one, winners perspective", Beats{nodeOneFeasible, nodeTwoFeasible}, nodeOneFeasible, true, false},
		{"two feasible nodes pass one, losers perspective", Beats{nodeOneFeasible, nodeTwoFeasible}, nodeTwoFeasible, false, false},
		{"two feasible nodes pass two, winners perspective", Beats{nodeOneCoordinator, nodeTwoNone}, nodeOneCoordinator, true, false},
		{"two feasible nodes pass two, losers perspective", Beats{nodeOneCoordinator, nodeTwoNone}, nodeTwoFeasible, false, true},
		{"two coordinator nodes, winners perspective", Beats{nodeOneCoordinator, nodeTwoCoordinator}, nodeOneCoordinator, true, false},
		{"two coordinator nodes, losers perspective", Beats{nodeOneCoordinator, nodeTwoCoordinator}, nodeTwoCoordinator, false, false},
	}

	for _, table := range tables {

		isCoord, isFeas := table.knownBeats.Evaluate(table.self.Coordinator, table.self.Feasible, table.self.ID)
		if isFeas != table.shouldBeFeas {
			t.Errorf("Error in Evaluate() for case \"%s\", Feasible was %t, should be %t", table.description, isFeas, table.shouldBeFeas)
		}
		if isCoord != table.shouldBeCoord {
			t.Errorf("Error in Evaluate() for case \"%s\", Coordinator was %t, should be %t", table.description, isCoord, table.shouldBeCoord)
		}
	}
}
