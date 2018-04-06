package heartbeat

// Evaluate assesses the set of known nodes to determine which node has/should have the Coordinator role
func (beats Beats) Evaluate() {
	if beats.CoordCount() == 0 {
		if Self.Coordinator {
			log.Infoln("This node is the uncontested coordinator, no further evaluation")
			Self.Feasible = false
			return
		} else {
			log.Infoln("No coordinators exist at all")
			if Self.Feasible {
				log.Infoln("This node is the feasible coordinator, promote")
				Self.Coordinator = true
				Self.Feasible = false
				return
			} else {
				log.Infoln("This node is not the feasible coordinator, assess feasible coordinator status")
				beats.evaluateFeas()
				return
			}
		}
	} else {
		if Self.Coordinator {
			best, _ := beats.bestActiveCoord()
			if best < node.Self.ID {
				log.Infoln("This node is not the best coordinator, update Self and assess feasible coordinators")
				Self.Coordinator = false
				beats.evaluateFeas()
				return
			} else {
				log.Infoln("node is the best coordinator, cease further evaluation")
				Self.Feasible = false
				return
			}
		} else {
			log.Infoln("There is one or more coordinators already, move to evaluating feasible coordinators")
			beats.evaluateFeas()
			return
		}
	}
}

// Evaluate assesses the set of known nodes to determine which node has/should have the Feasible Coordinator role
func (beats Beats) evaluateFeas() {
	if beats.FeasCount() == 0 {
		if Self.Feasible {
			log.Infoln("This node is the uncontested feasible coordinator, no further evaluation")
			return
		} else {
			log.Infoln("No feasible coordinators exist")
			best, _ := beats.bestFeas()
			if best < node.Self.ID {
				log.Infoln("This node is not the best possible feasible coordinator, unset")
				Self.Feasible = false
				return

			} else {
				log.Infoln("This node is the best possible feasible coordinator, set")
				Self.Feasible = true
				return
			}

		}
	} else {
		if Self.Feasible {
			log.Infoln("This node is a contested feasible coordinator")
			best, _ := beats.bestActiveFeas()
			if best < node.Self.ID {
				log.Infoln("This node is not the best feasible coordinator in contention, unset")
				// WARNING: This means incomplete/one-way message transmission may cause flapping!
				Self.Feasible = false
			}
		}
	}
}
