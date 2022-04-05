package order_redundancy

import (
	. "Elevator-project/src/constants"
	"Elevator-project/src/elevator_io"
	"time"
)

// Refer to the finite state machine diagram in README for a full overview of order state transitions

func OrderRedundancy(
	net_or_remoteOrders 	<-chan OrdersMSG,
	eio_or_buttons 			<-chan elevator_io.ButtonEvent,
	ec_or_localOrderServed 	<-chan elevator_io.ButtonEvent,
	al_or_newElevDetected 	<-chan string,
	al_or_elevsLost 		<-chan []string,
	al_or_disconnected 		<-chan bool,
	or_oa_confirmedOrders 	chan<- ConfirmedOrders,
	or_net_localOrders 		chan<- OrdersMSG,
	id string,
) {

	var alive_elevators []string

	var orders Orders
	orders.CabCalls = make(map[string]*[N_FLOORS]OrderState)
	orders.CabCalls[id] = &[N_FLOORS]OrderState{}

	orders.CabCallConsensus = make(map[string]*[N_FLOORS][]string)
	orders.CabCallConsensus[id] = &[N_FLOORS][]string{}

	io_setAllLights(id, orders)

	sendOrderTicker := time.NewTicker(PERIODIC_SEND_DURATION)

	for {
		select {
		case new_elev := <-al_or_newElevDetected:
			alive_elevators = append(alive_elevators, new_elev)
			// allocate space if elevator completely new
			if _, ok := orders.CabCalls[new_elev]; !ok {
				orders.CabCalls[new_elev] = &[N_FLOORS]OrderState{}
				orders.CabCallConsensus[new_elev] = &[N_FLOORS][]string{}
			}

		case lost_elevs := <-al_or_elevsLost:
			for _, lost_id := range lost_elevs {
				alive_elevators = remove(alive_elevators, lost_id)
				if id != lost_id {
					for floor := 0; floor < N_FLOORS; floor++ {
						// set cab call to OS_Unknown unless it is confirmed (redundancy)
						if orders.CabCalls[lost_id][floor] != OS_Confirmed {
							orders.CabCalls[lost_id][floor] = OS_Unknown
							orders.CabCallConsensus[lost_id][floor] = []string{}
						}
					}
				}
			}

		case <-al_or_disconnected:
			for floor := 0; floor < N_FLOORS; floor++ {
				for btn := 0; btn < N_BTN_TYPES-1; btn++ {
					orders.HallCallConsensus[floor][btn] = []string{}

					switch orders.HallCalls[floor][btn] {
					case OS_Unconfirmed, OS_None:
						orders.HallCalls[floor][btn] = OS_Unknown
					case OS_Confirmed:
					case OS_Unknown:
					}
				}

			}

		case remote_orders := <-net_or_remoteOrders:
			if !contains(alive_elevators, remote_orders.Id) {
				break
			}

			// define order state transitions (cyclic counter) as function
			orderStateMachine := func(
				local_order_ptr 	*OrderState,
				local_consensus_ptr *[]string,
				remote_order 		OrderState,
				remote_consensus 	[]string,
			) {
				switch remote_order {
				case OS_Unknown:
					switch *local_order_ptr {
					case OS_Unknown:
					case OS_Confirmed:
					case OS_None:
					case OS_Unconfirmed:
					}
				case OS_Confirmed:
					switch *local_order_ptr {
					case OS_Unknown:
						*local_order_ptr = remote_order
						*local_consensus_ptr = insert(*local_consensus_ptr, id)
						*local_consensus_ptr = insert(*local_consensus_ptr, remote_orders.Id)
						if consensusCheck(*local_consensus_ptr, remote_consensus, alive_elevators) {
							*local_order_ptr = OS_Confirmed
						}
					case OS_Confirmed:
					case OS_None:
					case OS_Unconfirmed:
						*local_consensus_ptr = insert(*local_consensus_ptr, remote_orders.Id)
						if consensusCheck(*local_consensus_ptr, remote_consensus, alive_elevators) {
							*local_order_ptr = OS_Confirmed
						}
					}
				case OS_None:
					switch *local_order_ptr {
					case OS_Unknown:
						*local_order_ptr = remote_order
					case OS_Confirmed:
						*local_order_ptr = remote_order
						*local_consensus_ptr = []string{}
					case OS_None:
					case OS_Unconfirmed:
					}
				case OS_Unconfirmed:
					switch *local_order_ptr {
					case OS_Unknown, OS_None, OS_Unconfirmed:
						*local_order_ptr = remote_order
						*local_consensus_ptr = insert(*local_consensus_ptr, id)
						*local_consensus_ptr = insert(*local_consensus_ptr, remote_orders.Id)
						if consensusCheck(*local_consensus_ptr, remote_consensus, alive_elevators) {
							*local_order_ptr = OS_Confirmed
						}

					case OS_Confirmed:
					}
				}
			}

			// Perform state transitions on hall calls
			if len(alive_elevators) > 1 {
				for floor := 0; floor < N_FLOORS; floor++ {
					for btn := 0; btn < 2; btn++ {
						orderStateMachine(
							&orders.HallCalls[floor][btn],
							&orders.HallCallConsensus[floor][btn],
							remote_orders.HallCalls[floor][btn],
							remote_orders.HallCallConsensus[floor][btn])
					}
				}
			}

			// Perform state transitions on cab calls
			for elev_id := range orders.CabCalls {
				for floor := 0; floor < N_FLOORS; floor++ {
					orderStateMachine(
						&orders.CabCalls[elev_id][floor],
						&orders.CabCallConsensus[elev_id][floor],
						remote_orders.CabCalls[elev_id][floor],
						remote_orders.CabCallConsensus[elev_id][floor])
				}
			}

		case button_event := <-eio_or_buttons:
			floor := button_event.Floor
			btn := button_event.Button
			switch btn {
			case elevator_io.BT_Cab:
				if orders.CabCalls[id][floor] != OS_Confirmed {
					if len(alive_elevators) <= 1 {
						orders.CabCalls[id][floor] = OS_Confirmed
					} else {
						orders.CabCalls[id][floor] = OS_Unconfirmed
					}

				}
			case elevator_io.BT_HallUp, elevator_io.BT_HallDown:
				if len(alive_elevators) > 1 {
					if orders.HallCalls[floor][btn] != OS_Confirmed {
						orders.HallCalls[floor][btn] = OS_Unconfirmed
						orders.HallCallConsensus[floor][btn] = insert(orders.HallCallConsensus[floor][btn], id)
					}
				}
			}

		case served_order := <-ec_or_localOrderServed:
			floor := served_order.Floor
			btn := served_order.Button
			switch btn {
			case elevator_io.BT_Cab:
				if orders.CabCalls[id][floor] == OS_Confirmed {
					orders.CabCalls[id][floor] = OS_None
					orders.CabCallConsensus[id][floor] = []string{}
				}
			case elevator_io.BT_HallUp, elevator_io.BT_HallDown:
				if len(alive_elevators) > 1 {
					orders.HallCalls[floor][btn] = OS_None
					orders.HallCallConsensus[floor][btn] = []string{}
				} else {
					orders.HallCalls[floor][btn] = OS_Unknown
					orders.HallCallConsensus[floor][btn] = []string{}
				}
			}

		case <-sendOrderTicker.C:
			orders_msg := createOrdersMSG(id, orders)
			or_net_localOrders <- orders_msg
			confirmed_orders := getConfirmedOrders(id, orders, alive_elevators)
			or_oa_confirmedOrders <- confirmed_orders
			io_setAllLights(id, orders)
		}
	}
}

func io_setAllLights(id string, orders Orders) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BTN_TYPES; btn++ {
			is_order := false
			switch elevator_io.ButtonType(btn) {
			case elevator_io.BT_Cab:
				if orders.CabCalls[id][floor] == OS_Confirmed {
					is_order = true
				}
			case elevator_io.BT_HallDown, elevator_io.BT_HallUp:
				if orders.HallCalls[floor][btn] == OS_Confirmed {
					is_order = true
				}
			}
			elevator_io.SetButtonLamp(elevator_io.ButtonType(btn), floor, is_order)
		}
	}
}

func consensusCheck(local_consensus []string, remote_consensus []string, alive_elevators []string) bool {
	for _, elev_id := range alive_elevators {
		if !contains(local_consensus, elev_id) || !contains(remote_consensus, elev_id) {
			return false
		}
	}
	return true
}

//Helper functions

func insert(str_list []string, str string) []string {
	if !contains(str_list, str) {
		return append(str_list, str)
	}
	return str_list
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
