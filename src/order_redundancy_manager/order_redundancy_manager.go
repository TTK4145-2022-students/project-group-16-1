package order_redundancy_manager

import (
	"Elevator-project/src/elevio"
	"time"
)

const N_FLOORS = 4 //REMOVE THIS
const N_BTN_TYPES = 3
const HARDWARE_ADDR = "localhost:15657"
const INTERVAL = 45 * time.Millisecond

type OrderState int

const (
	OS_Unknown     OrderState = 0
	OS_Confirmed   OrderState = 1
	OS_None        OrderState = 2
	OS_Unconfirmed OrderState = 3
)

type Orders struct {
	Calls map[string]*[N_BTN_TYPES][N_FLOORS]OrderState
}

type OrdersMSG struct {
	Id    string
	Calls map[string][N_BTN_TYPES][N_FLOORS]OrderState
	// HallCalls [N_FLOORS][2]OrderState
	// CabCalls  map[string][N_FLOORS]OrderState
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

	orders.Calls = make(map[string]*[N_BTN_TYPES][N_FLOORS]OrderState)
	orders.Calls[id] = &[N_BTN_TYPES][N_FLOORS]OrderState{}

	periodic_timer := time.After(INTERVAL)

	for {
		select {
		case new_elev := <-al_or_newElevDetected:
			alive_elevators = append(alive_elevators, new_elev)
			orders.Calls[new_elev] = &[N_BTN_TYPES][N_FLOORS]OrderState{}
		case lost_elev_ids := <-al_or_elevsLost:
			for i := range lost_elev_ids {
				alive_elevators = remove(alive_elevators, lost_elev_ids[i])
				// Set cab call state to unknown if not confirmed
				for btn := 0; btn < N_BTN_TYPES; btn++ {
					for floor := 0; floor < N_FLOORS; floor++ {
						if orders.Calls[lost_elev_ids[i]][btn][floor] != OS_Confirmed {
							orders.Calls[lost_elev_ids[i]][btn][floor] = OS_Unknown
						}
					}
				}
			}
		case <-al_or_disconnected:
			for _, btn := range []elevio.ButtonType{elevio.BT_HallDown, elevio.BT_HallUp} {
				for floor := 0; floor < N_FLOORS; floor++ {
					if orders.Calls[id][btn][floor] == OS_None || orders.Calls[id][btn][floor] == OS_Unconfirmed {
						orders.Calls[id][btn][floor] = OS_Unknown
					}
				}
			}

		case remote_orders := <-net_or_remoteOrders:
			// Check if elevator is alive
			if !contains(alive_elevators, remote_orders.Id) {
				break
			}

			new_confirmed_order := false

			for btn := 0; btn < N_BTN_TYPES; btn++ {
				for floor := 0; floor < N_FLOORS; floor++ {
					if btn == elevio.BT_Cab {
						if orders.Calls[remote_orders.Id][btn][floor] < remote_orders.Calls[remote_orders.Id][btn][floor] {
							orders.Calls[remote_orders.Id][btn][floor] = remote_orders.Calls[remote_orders.Id][btn][floor]
							if barrierCheck(remote_orders.Id, floor, btn, orders, alive_elevators) {
								orders.Calls[remote_orders.Id][floor][btn] = OS_Confirmed
								new_confirmed_order = true
							}
						}
					} else {
						if orders.Calls[id][btn][floor] < remote_orders.Calls[remote_orders.Id][btn][floor] {
							orders.Calls[id][btn][floor] = remote_orders.Calls[remote_orders.Id][btn][floor]
							orders.Calls[remote_orders.Id][btn][floor] = remote_orders.Calls[remote_orders.Id][btn][floor]
							if barrierCheck(remote_orders.Id, floor, btn, orders, alive_elevators) {
								orders.Calls[id][floor][btn] = OS_Confirmed
								new_confirmed_order = true
							}
						}
					}
				}
			}

			if new_confirmed_order {
				confirmed_orders := getConfirmedOrders(id, orders, alive_elevators)
				or_oa_confirmedOrders <- confirmed_orders
			}
		case button_event := <-drv_or_buttons:

			floor := button_event.Floor
			switch button_event.Button {
			case elevio.BT_Cab:
				if orders.Calls[id][elevio.BT_Cab][floor] != OS_Confirmed {
					orders.Calls[id][elevio.BT_Cab][floor] = OS_Unconfirmed

				}
			default:
				if len(alive_elevators) > 1 {
					if orders.Calls[id][button_event.Button][floor] != OS_Confirmed {
						orders.Calls[id][button_event.Button][floor] = OS_Unconfirmed
					}
				}
			}
		case served_order := <-ec_or_localOrderServed:
			floor := served_order.Floor
			btn := served_order.Button
			switch btn {
			case elevio.BT_Cab:
				orders.Calls[id][elevio.BT_Cab][floor] = OS_None
			default:
				if len(alive_elevators) > 1 {
					orders.Calls[id][btn][floor] = OS_None
				} else {
					orders.Calls[id][btn][floor] = OS_Unknown
				}
			}
		case <-periodic_timer:
			orders_msg := createOrdersMSG(id, orders)
			or_net_localOrders <- orders_msg
			periodic_timer = time.After(INTERVAL)
		}
	}
}

func createOrdersMSG(id string, orders Orders) OrdersMSG {
	var orders_msg OrdersMSG
	orders_msg.Id = id
	orders_msg.Calls = make(map[string][N_BTN_TYPES][N_FLOORS]OrderState)
	for id, val := range orders.Calls {
		orders_msg.Calls[id] = *val
	}
	return orders_msg
}

func getConfirmedOrders(id string, orders Orders, alive_elevators []string) ConfirmedOrders {
	var confirmed_orders ConfirmedOrders
	confirmed_orders.CabCalls = make(map[string]*[N_FLOORS]bool)

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++ {
			if orders.Calls[id][btn][floor] == OS_Confirmed {
				confirmed_orders.HallCalls[floor][btn] = true
			}
		}
	}
	for _, elev_id := range alive_elevators {
		confirmed_orders.CabCalls[elev_id] = &[N_FLOORS]bool{}
		for floor := 0; floor < N_FLOORS; floor++ {
			if orders.Calls[elev_id][elevio.BT_Cab][floor] == OS_Confirmed {
				confirmed_orders.CabCalls[elev_id][floor] = true
			}
		}
	}
	return confirmed_orders
}

func barrierCheck(id string, floor int, btn int, orders Orders, alive_elevators []string) bool {
	for _, elev_id := range alive_elevators {
		if orders.Calls[elev_id][btn][floor] != OS_Unconfirmed {
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
