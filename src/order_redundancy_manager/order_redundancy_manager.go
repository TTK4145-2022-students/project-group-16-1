package order_redundancy_manager

import (
	"Elevator-project/src/elevio"
	"time"
)

const N_FLOORS = 4 //REMOVE THIS
const N_BUTTONS = 3
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
	Id                string
	HallCalls         [N_FLOORS][2]OrderState
	HallCallConsensus map[string]*[N_FLOORS][2]bool
	CabCalls          map[string]*[N_FLOORS]OrderState
}

type ConfirmedOrders struct {
	HallCalls [N_FLOORS][2]bool
	CabCalls  map[string]*[N_FLOORS]bool
}

func OrderRedundancyManager(
	orm_remoteOrders <-chan Orders,
	drv_buttons <-chan elevio.ButtonEvent,
	ec_localOrderServed <-chan elevio.ButtonEvent,
	orm_newElevDetected <-chan string,
	orm_elevsLost <-chan []string,
	orm_disconnected <-chan bool,
	orm_confirmedOrders chan<- ConfirmedOrders,
	orm_localOrders chan<- Orders,
	id string) {

	var alive_elevators []string
	var orders Orders
	var connected bool

	orders.CabCalls = make(map[string]*[N_FLOORS]OrderState)
	orders.CabCalls[id] = &[N_FLOORS]OrderState{}
	orders.HallCallConsensus = make(map[string]*[N_FLOORS][2]bool)
	orders.Id = id

	timeout := time.After(INTERVAL)

	for {
		select {
		case new_elev := <-orm_newElevDetected:
			alive_elevators = append(alive_elevators, new_elev)
			orders.HallCallConsensus[new_elev] = &[N_FLOORS][2]bool{}
			connected = true
		case lost_elevs := <-orm_elevsLost:
			for i := range lost_elevs {
				alive_elevators = remove(alive_elevators, lost_elevs[i])
				// Set cabcall state to unknown if not confirmed
				for floor := 0; floor < N_FLOORS; floor++ {
					if orders.CabCalls[lost_elevs[i]][floor] != OS_Confirmed {
						orders.CabCalls[lost_elevs[i]][floor] = OS_Unknown
					}
				}
				// Reset hallcall consensus matrix for given ID
				for id := range orders.HallCallConsensus {
					orders.HallCallConsensus[id] = &[N_FLOORS][2]bool{}
				}
			}
		case disconnected := <-orm_disconnected:
			connected = !disconnected
			for floor := 0; floor < N_FLOORS; floor++ {
				orders.HallCalls[floor][0] = fsm_onDisconnect(orders.HallCalls[floor][0])
				orders.HallCalls[floor][1] = fsm_onDisconnect(orders.HallCalls[floor][1])
			}
		case remoteOrders := <-orm_remoteOrders:
			// Check if elevator is alive

			if !contains(alive_elevators, remoteOrders.Id) {
				break
			}
			new_confirmed_order := false

			// Loop through all HC orders
			for floor := 0; floor < N_FLOORS; floor++ {
				for btn := 0; btn < 2; btn++ {
					orders.HallCalls[floor][btn] = fsm_remoteReceived(orders.HallCalls[floor][btn], remoteOrders.HallCalls[floor][btn])
					if orders.HallCalls[floor][btn] == OS_Unconfirmed {
						if fsm_barrierCheck(remoteOrders.Id, floor, btn, orders, alive_elevators) {
							orders.HallCalls[floor][btn] = OS_Confirmed
							new_confirmed_order = true
						}
					}
				}
			}
			// Loop through all CC orders
			for elev_id := range orders.CabCalls {
				for floor := 0; floor < N_FLOORS; floor++ {
					orders.CabCalls[elev_id][floor] = fsm_remoteReceived(orders.CabCalls[elev_id][floor], remoteOrders.CabCalls[elev_id][floor])
					if orders.CabCalls[elev_id][floor] == OS_Unconfirmed {
						if fsm_barrierCheck(remoteOrders.Id, floor, 2, orders, alive_elevators) {
							orders.CabCalls[elev_id][floor] = OS_Confirmed
							new_confirmed_order = true
						}
					}
				}

			}

			if new_confirmed_order {
				confirmed_orders := fsm_getConfirmedOrders(orders, alive_elevators)
				orm_confirmedOrders <- confirmed_orders
			}
		case button_event := <-drv_buttons:

			floor := button_event.Floor
			switch button_event.Button {
			case elevio.BT_Cab:

				if orders.CabCalls[orders.Id][floor] != OS_Confirmed {
					orders.CabCalls[orders.Id][floor] = OS_Unconfirmed

				}
			case elevio.BT_HallUp:
				if connected {
					if orders.HallCalls[floor][0] != OS_Confirmed {
						orders.HallCalls[floor][0] = OS_Unconfirmed
					}
				}
			case elevio.BT_HallDown:
				if connected {
					if orders.HallCalls[floor][1] != OS_Confirmed {
						orders.HallCalls[floor][1] = OS_Unconfirmed
					}
				}
			}
		case served_order := <-ec_localOrderServed:
			floor := served_order.Floor
			btn := served_order.Button
			switch btn {
			case elevio.BT_Cab:
				orders.CabCalls[id][floor] = OS_None
			case elevio.BT_HallUp:
				if connected {
					orders.HallCalls[floor][0] = OS_None
				} else {
					orders.HallCalls[floor][0] = OS_Unknown
				}
			case elevio.BT_HallDown:
				if connected {
					orders.HallCalls[floor][1] = OS_None
				} else {
					orders.HallCalls[floor][1] = OS_Unknown
				}
			}
		case <-timeout:
			orm_localOrders <- orders
			timeout = time.After(INTERVAL)
		}
	}
}

func fsm_getConfirmedOrders(orders Orders, alive_elevators []string) ConfirmedOrders {
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
	return confirmed_orders
}
func fsm_barrierCheck(id string, floor int, btn int, orders Orders, alive_elevators []string) bool {

	if len(alive_elevators) == 1 {
		if btn == 2 && id == orders.Id {
			return true
		} else {
			return false
		}
	}
	if btn == 2 {
		for _, elev_id := range alive_elevators {
			if orders.CabCalls[elev_id][floor] != OS_Confirmed {
				return false
			}
		}
	} else {
		orders.HallCallConsensus[id][floor][btn] = true
		for _, elev_id := range alive_elevators {
			if !orders.HallCallConsensus[elev_id][floor][btn] {
				return false
			}
		}
	}
	return true
}

func fsm_onDisconnect(order_state OrderState) OrderState {
	switch order_state {
	case OS_Unconfirmed, OS_None:
		return OS_Unknown
	default:
		return order_state
	}
}

func fsm_remoteReceived(local_HC OrderState, remote_HC OrderState) OrderState {
	if remote_HC > local_HC {
		return remote_HC
	} else {
		return local_HC
	}

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
