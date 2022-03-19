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
	orm_remoteOrders <-chan OrdersMSG,
	drv_buttons <-chan elevio.ButtonEvent,
	ec_localOrderServed <-chan elevio.ButtonEvent,
	orm_newElevDetected <-chan string,
	orm_elevsLost <-chan []string,
	orm_disconnected <-chan bool,
	orm_confirmedOrders chan<- ConfirmedOrders,
	orm_localOrders chan<- OrdersMSG,
	id string) {

	var alive_elevators []string
	var orders Orders

	orders.CabCalls = make(map[string]*[N_FLOORS]OrderState)
	orders.CabCalls[id] = &[N_FLOORS]OrderState{}

	orders.CabCallConsensus = make(map[string]*[N_FLOORS][]string)
	orders.CabCallConsensus[id] = &[N_FLOORS][]string{}

	io_setAllLights(id, orders)

	periodicTimeout := time.After(INTERVAL)

	for {
		select {
		case new_elev := <-orm_newElevDetected:
			alive_elevators = append(alive_elevators, new_elev)
			if _, ok := orders.CabCalls[new_elev]; !ok {
				orders.CabCalls[new_elev] = &[N_FLOORS]OrderState{}
				orders.CabCallConsensus[new_elev] = &[N_FLOORS][]string{}
			}
		case lost_elevs := <-orm_elevsLost:
			for i := range lost_elevs {
				alive_elevators = remove(alive_elevators, lost_elevs[i])
				// Set cabcall state to unknown if not confirmed
				for floor := 0; floor < N_FLOORS; floor++ {
					if orders.CabCalls[lost_elevs[i]][floor] != OS_Confirmed {
						orders.CabCalls[lost_elevs[i]][floor] = OS_Unknown
					}
				}
			}
		case <-orm_disconnected:
			for floor := 0; floor < N_FLOORS; floor++ {
				for btn := 0; btn < 2; btn++ {
					switch orders.HallCalls[floor][btn] {
					case OS_Unconfirmed, OS_None:
						orders.HallCalls[floor][btn] = OS_Unknown
					case OS_Confirmed:
					case OS_Unknown:
					}
				}

			}
		case remote_orders := <-orm_remoteOrders:
			// Check if elevator is alive
			if !contains(alive_elevators, remote_orders.Id) {
				break
			}
			for floor := 0; floor < N_FLOORS; floor++ {
				for btn := 0; btn < 2; btn++ {
					if orders.HallCalls[floor][btn] <= remote_orders.HallCalls[floor][btn] {
						orders.HallCalls[floor][btn] = remote_orders.HallCalls[floor][btn]
						if remote_orders.HallCalls[floor][btn] == OS_Unconfirmed {
							if !contains(orders.HallCallConsensus[floor][btn], remote_orders.Id) {
								orders.HallCallConsensus[floor][btn] = append(orders.HallCallConsensus[floor][btn], remote_orders.Id)
							}
						}
					}
					//Consensus check
					if orders.HallCalls[floor][btn] == OS_Unconfirmed && len(alive_elevators) > 1 {
						if consensusCheck(orders.HallCallConsensus[floor][btn], remote_orders.HallCallConsensus[floor][btn], alive_elevators) {
							orders.HallCalls[floor][btn] = OS_Confirmed
							orders.HallCallConsensus[floor][btn] = []string{}
						}
					}
				}
			}

			for elev_id := range orders.CabCalls {
				for floor := 0; floor < N_FLOORS; floor++ {
					if orders.CabCalls[elev_id][floor] <= remote_orders.CabCalls[elev_id][floor] {
						orders.CabCalls[elev_id][floor] = remote_orders.CabCalls[elev_id][floor]
						if remote_orders.CabCalls[elev_id][floor] == OS_Unconfirmed {
							if !contains(orders.CabCallConsensus[elev_id][floor], remote_orders.Id) {
								orders.CabCallConsensus[elev_id][floor] = append(orders.CabCallConsensus[elev_id][floor], remote_orders.Id)
							}
						}
					}
					//Consensus check
					if orders.CabCalls[elev_id][floor] == OS_Unconfirmed {
						if consensusCheck(orders.CabCallConsensus[elev_id][floor], remote_orders.CabCallConsensus[elev_id][floor], alive_elevators) {
							orders.CabCalls[elev_id][floor] = OS_Confirmed
							orders.CabCallConsensus[elev_id][floor] = []string{}
						}
					}

				}

			}

		case button_event := <-drv_buttons:
			floor := button_event.Floor
			btn := button_event.Button
			switch btn {
			case elevio.BT_Cab:
				if orders.CabCalls[id][floor] != OS_Confirmed {
					if len(alive_elevators) == 0 {
						orders.CabCalls[id][floor] = OS_Confirmed
					} else {
						orders.CabCalls[id][floor] = OS_Unconfirmed
					}

				}
			case elevio.BT_HallUp, elevio.BT_HallDown:
				if len(alive_elevators) > 1 {
					if orders.HallCalls[floor][btn] != OS_Confirmed {
						orders.HallCalls[floor][btn] = OS_Unconfirmed
					}
				}

			}

		case served_order := <-ec_localOrderServed:
			floor := served_order.Floor
			btn := served_order.Button
			switch btn {
			case elevio.BT_Cab:
				if orders.CabCalls[id][floor] == OS_Confirmed {
					orders.CabCalls[id][floor] = OS_None
				}
			case elevio.BT_HallUp, elevio.BT_HallDown:
				if len(alive_elevators) > 1 {
					orders.HallCalls[floor][btn] = OS_None
				} else {
					orders.HallCalls[floor][btn] = OS_Unknown
				}
			}

		case <-periodicTimeout:
			orders_msg := createOrdersMSG(id, orders)
			orm_localOrders <- orders_msg
			confirmed_orders := getConfirmedOrders(id, orders, alive_elevators)
			orm_confirmedOrders <- confirmed_orders
			io_setAllLights(id, orders)
			periodicTimeout = time.After(INTERVAL)
		}
	}
}

func io_setAllLights(id string, orders Orders) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
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
