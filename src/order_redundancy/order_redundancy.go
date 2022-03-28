package order_redundancy

import (
	. "Elevator-project/src/constants"
	"Elevator-project/src/elevio"
	"time"
)

type OrderState int

const (
	OS_Unknown     OrderState = 0
	OS_Confirmed   OrderState = 1
	OS_None        OrderState = 2
	OS_Unconfirmed OrderState = 3
)

type Orders struct {
	HallCalls         [N_FLOORS][2]OrderState
	HallCallConsensus [N_FLOORS][2][]string
	CabCalls          map[string]*[N_FLOORS]OrderState
	CabCallConsensus  map[string]*[N_FLOORS][]string
}

type OrdersMSG struct {
	Id                string
	HallCalls         [N_FLOORS][2]OrderState
	HallCallConsensus [N_FLOORS][2][]string
	CabCalls          map[string][N_FLOORS]OrderState
	CabCallConsensus  map[string][N_FLOORS][]string
}

type ConfirmedOrders struct {
	HallCalls [N_FLOORS][2]bool
	CabCalls  map[string]*[N_FLOORS]bool
}

func OrderRedundancyManager(
	net_or_remoteOrders <-chan OrdersMSG,
	drv_or_buttons <-chan elevio.ButtonEvent,
	ec_or_localOrderServed <-chan elevio.ButtonEvent,
	al_or_newElevDetected <-chan string,
	al_or_elevsLost <-chan []string,
	al_or_disconnected <-chan bool,
	or_oa_confirmedOrders chan<- ConfirmedOrders,
	or_net_localOrders chan<- OrdersMSG,
	id string) {

	var alive_elevators []string
	var orders Orders

	orders.CabCalls = make(map[string]*[N_FLOORS]OrderState)
	orders.CabCalls[id] = &[N_FLOORS]OrderState{}

	orders.CabCallConsensus = make(map[string]*[N_FLOORS][]string)
	orders.CabCallConsensus[id] = &[N_FLOORS][]string{}

	io_setAllLights(id, orders)

	periodicTimeout := time.NewTicker(PERIODIC_SEND_DURATION)

	for {
		select {
		case new_elev := <-al_or_newElevDetected:
			alive_elevators = append(alive_elevators, new_elev)
			if _, ok := orders.CabCalls[new_elev]; !ok {
				orders.CabCalls[new_elev] = &[N_FLOORS]OrderState{}
				orders.CabCallConsensus[new_elev] = &[N_FLOORS][]string{}
			}
		case lost_elevs := <-al_or_elevsLost:
			for i, lost_id := range lost_elevs {
				alive_elevators = remove(alive_elevators, lost_elevs[i])
				// Set cabcall state to unknown if not confirmed, but not own cabcalls
				if id == lost_id {
					break
				}
				for floor := 0; floor < N_FLOORS; floor++ {
					if orders.CabCalls[lost_elevs[i]][floor] != OS_Confirmed {
						orders.CabCalls[lost_elevs[i]][floor] = OS_Unknown
						orders.CabCallConsensus[lost_elevs[i]][floor] = []string{}
					}
				}
			}
		case <-al_or_disconnected:
			for floor := 0; floor < N_FLOORS; floor++ {
				for btn := 0; btn < 2; btn++ {
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
			// Check if elevator is alive
			if !contains(alive_elevators, remote_orders.Id) {
				break
			}

			order_fsm := func(local_order_ptr *OrderState,
				local_consensus_ptr *[]string,
				remote_order OrderState,
				remote_consensus []string) {
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
						*local_consensus_ptr = addIdToConsensus(*local_consensus_ptr, id)
						*local_consensus_ptr = addIdToConsensus(*local_consensus_ptr, remote_orders.Id)
						if consensusCheck(*local_consensus_ptr, remote_consensus, alive_elevators) {
							*local_order_ptr = OS_Confirmed
						}
					case OS_Confirmed:
					case OS_None:
					case OS_Unconfirmed:
						*local_consensus_ptr = addIdToConsensus(*local_consensus_ptr, remote_orders.Id)
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
						*local_consensus_ptr = addIdToConsensus(*local_consensus_ptr, id)
						*local_consensus_ptr = addIdToConsensus(*local_consensus_ptr, remote_orders.Id)
						if consensusCheck(*local_consensus_ptr, remote_consensus, alive_elevators) {
							*local_order_ptr = OS_Confirmed
						}

					case OS_Confirmed:
					}

				}
			}
			// Loop through hallcalls
			if len(alive_elevators) > 1 {
				for floor := 0; floor < N_FLOORS; floor++ {
					for btn := 0; btn < 2; btn++ {
						local_order_ptr := &orders.HallCalls[floor][btn]
						local_consensus_ptr := &orders.HallCallConsensus[floor][btn]
						remote_order := remote_orders.HallCalls[floor][btn]
						remote_consensus := remote_orders.HallCallConsensus[floor][btn]
						order_fsm(local_order_ptr, local_consensus_ptr, remote_order, remote_consensus)
					}
				}
			}
			// Loop through cabcalls
			for elev_id := range orders.CabCalls {
				for floor := 0; floor < N_FLOORS; floor++ {
					local_order_ptr := &orders.CabCalls[elev_id][floor]
					local_consensus_ptr := &orders.CabCallConsensus[elev_id][floor]
					remote_order := remote_orders.CabCalls[elev_id][floor]
					remote_consensus := remote_orders.CabCallConsensus[elev_id][floor]
					order_fsm(local_order_ptr, local_consensus_ptr, remote_order, remote_consensus)
				}
			}

		case button_event := <-drv_or_buttons:
			floor := button_event.Floor
			btn := button_event.Button
			switch btn {
			case elevio.BT_Cab:
				if orders.CabCalls[id][floor] != OS_Confirmed {
					if len(alive_elevators) < 2 {
						orders.CabCalls[id][floor] = OS_Confirmed
					} else {
						orders.CabCalls[id][floor] = OS_Unconfirmed
					}

				}
			case elevio.BT_HallUp, elevio.BT_HallDown:
				if len(alive_elevators) > 1 {
					if orders.HallCalls[floor][btn] != OS_Confirmed {
						orders.HallCalls[floor][btn] = OS_Unconfirmed
						orders.HallCallConsensus[floor][btn] = addIdToConsensus(orders.HallCallConsensus[floor][btn], id)
					}
				}

			}

		case served_order := <-ec_or_localOrderServed:
			floor := served_order.Floor
			btn := served_order.Button
			switch btn {
			case elevio.BT_Cab:
				if orders.CabCalls[id][floor] == OS_Confirmed {
					orders.CabCalls[id][floor] = OS_None
					orders.CabCallConsensus[id][floor] = []string{}
				}
			case elevio.BT_HallUp, elevio.BT_HallDown:
				if len(alive_elevators) > 1 {
					orders.HallCalls[floor][btn] = OS_None
					orders.HallCallConsensus[floor][btn] = []string{}
				} else {
					orders.HallCalls[floor][btn] = OS_Unknown
					orders.HallCallConsensus[floor][btn] = []string{}
				}
			}

		case <-periodicTimeout.C:
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
			switch elevio.ButtonType(btn) {
			case elevio.BT_Cab:
				if orders.CabCalls[id][floor] == OS_Confirmed {
					is_order = true
				}
			case elevio.BT_HallDown, elevio.BT_HallUp:
				if orders.HallCalls[floor][btn] == OS_Confirmed {
					is_order = true
				}
			}
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, is_order)
		}
	}
}

func createOrdersMSG(id string, orders Orders) OrdersMSG {
	var orders_msg OrdersMSG
	orders_msg.Id = id
	orders_msg.CabCalls = make(map[string][N_FLOORS]OrderState)
	orders_msg.HallCalls = orders.HallCalls
	orders_msg.HallCallConsensus = orders.HallCallConsensus
	orders_msg.CabCallConsensus = make(map[string][N_FLOORS][]string)
	for id, val := range orders.CabCalls {
		orders_msg.CabCalls[id] = *val
	}
	for id, val := range orders.CabCallConsensus {
		orders_msg.CabCallConsensus[id] = *val
	}
	return orders_msg
}

func getConfirmedOrders(id string, orders Orders, alive_elevators []string) ConfirmedOrders {
	var confirmed_orders ConfirmedOrders
	confirmed_orders.CabCalls = make(map[string]*[N_FLOORS]bool)

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++ {
			if orders.HallCalls[floor][btn] == OS_Confirmed {
				confirmed_orders.HallCalls[floor][btn] = true
			}
		}
	}
	for _, elev_id := range alive_elevators {
		confirmed_orders.CabCalls[elev_id] = &[N_FLOORS]bool{}
		for floor := 0; floor < N_FLOORS; floor++ {
			if orders.CabCalls[elev_id][floor] == OS_Confirmed {
				confirmed_orders.CabCalls[elev_id][floor] = true
			}
		}
	}
	confirmed_orders.CabCalls[id] = &[N_FLOORS]bool{}
	for floor := 0; floor < N_FLOORS; floor++ {

		if orders.CabCalls[id][floor] == OS_Confirmed {

			confirmed_orders.CabCalls[id][floor] = true
		}
	}
	return confirmed_orders
}
func consensusCheck(local_consensus []string, remote_consensus []string, alive_elevators []string) bool {
	for _, elev_id := range alive_elevators {
		if !contains(local_consensus, elev_id) || !contains(remote_consensus, elev_id) {
			return false
		}
	}
	return true
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

func addIdToConsensus(consensus []string, remote_id string) []string {
	if !contains(consensus, remote_id) {
		return append(consensus, remote_id)
	}
	return consensus
}
