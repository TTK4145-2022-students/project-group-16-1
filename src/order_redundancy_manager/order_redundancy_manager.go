package order_redundancy_manager

import (
	"Elevator-project/src/elevio"
)

const N_FLOORS = 4 //REMOVE THIS
const N_BUTTONS = 3
const HARDWARE_ADDR = "localhost:15657"

var alive_elevators map[string]bool
var num_alive_elevators int
var orders Orders

type OrderState int

const (
	OS_Unknown OrderState = iota
	OS_Confirmed
	OS_None
	OS_Unconfirmed
)

type Orders struct {
	hallCalls         [N_FLOORS][2]OrderState
	hallCallConsensus [N_FLOORS][2][]string
	cabCalls          map[string]*[N_FLOORS]OrderState
	cabCallConsensus  map[string]*[N_FLOORS][]string
}

type ConfirmedOrders struct {
	HallCalls [N_FLOORS][2]bool
	CabCalls  map[string][N_FLOORS]bool
}

func OrderRedundancyManager(orm_confirmedOrders chan ConfirmedOrders,
	orm_remoteOrders chan Orders,
	osr_aliveElevators chan map[string]bool,
	orm_localOrders chan Orders, drv_buttons chan elevio.ButtonEvent,
	ec_localOrderServed chan elevio.ButtonEvent, id string) {

	num_alive_elevators = 1
	alive_elevators = make(map[string]bool)
	alive_elevators[id] = true

	orders.cabCalls = make(map[string]*[N_FLOORS]OrderState)
	orders.cabCalls[id] = &[N_FLOORS]OrderState{}
	for {
		select {
		case alive_elevators = <-osr_aliveElevators:
			fsm_aliveElevatorsRecieved(id)
		case remoteOrders := <-orm_remoteOrders:
			fsm_remoteOrdersRecieved(remoteOrders, id)
		case button_event := <-drv_buttons:
			fsm_buttonPressed(button_event, id)
		case served_order := <-ec_localOrderServed:
			fsm_localOrderServed(served_order, id)
		}
	}
}

func fsm_buttonPressed(button_event elevio.ButtonEvent, id string) {
	floor := button_event.Floor
	switch button_event.Button {
	case elevio.BT_Cab:
		switch orders.cabCalls[id][floor] {
		case OS_Unknown, OS_None:
			orders.cabCalls[id][floor] = OS_Unconfirmed
		}
	case elevio.BT_HallUp:
		if num_alive_elevators > 1 {
			switch orders.hallCalls[floor][0] {
			case OS_Unknown, OS_None:
				orders.hallCalls[floor][0] = OS_Unconfirmed
			}
		}
	case elevio.BT_HallDown:
		if num_alive_elevators > 1 {
			switch orders.hallCalls[floor][1] {
			case OS_Unknown, OS_None:
				orders.hallCalls[floor][1] = OS_Unconfirmed
			}
		}
	}
}

func fsm_remoteOrdersRecieved(remoteOrders Orders, local_id string) {
	for id, _ := range remoteOrders.cabCalls {
		for floor := 0; floor < N_FLOORS; floor++ {
			if remoteOrders.cabCalls[id][floor] == OS_Unconfirmed {
				if !contains(orders.cabCallConsensus[id][floor], id) {
					orders.cabCallConsensus[id][floor] = append(orders.cabCallConsensus[id][floor], id)
				}
			}

			switch orders.cabCalls[id][floor] {
			case OS_Unknown:
				orders.cabCalls[id][floor] = remoteOrders.cabCalls[id][floor]
			case OS_Confirmed:
				if remoteOrders.cabCalls[id][floor] != OS_Unknown {
					orders.cabCalls[id][floor] = remoteOrders.cabCalls[id][floor]
				}
			case OS_None:
				if remoteOrders.cabCalls[id][floor] == OS_Unconfirmed {
					orders.cabCalls[id][floor] = remoteOrders.cabCalls[id][floor]
				}
			case OS_Unconfirmed:
				if len(orders.cabCallConsensus[id][floor]) == num_alive_elevators {
					if num_alive_elevators != 1 {
						orders.cabCalls[id][floor] = OS_Confirmed
					} else {
						if local_id == id {
							orders.cabCalls[id][floor] = OS_Confirmed
						}
					}
				}
			}
		}
	}

	for floor := 0; floor < N_FLOORS; floor++ {
		switch orders.hallCalls[floor][0] {
		case OS_Unknown:
			orders.hallCalls[floor][0] = remoteOrders.hallCalls[floor][0]
		case OS_Confirmed:
			if remoteOrders.hallCalls[floor][0] != OS_Unknown {
				orders.hallCalls[floor][0] = remoteOrders.hallCalls[floor][0]
			}
		case OS_None:
			if remoteOrders.hallCalls[floor][0] == OS_Unconfirmed {
				orders.hallCalls[floor][0] = remoteOrders.hallCalls[floor][0]
			}
		case OS_Unconfirmed:
			if len(orders.hallCallConsensus[floor][0]) == num_alive_elevators {
				orders.hallCalls[floor][0] = OS_Confirmed
			}
		}

		switch orders.hallCalls[floor][1] {
		case OS_Unknown:
			orders.hallCalls[floor][1] = remoteOrders.hallCalls[floor][1]
		case OS_Confirmed:
			if remoteOrders.hallCalls[floor][1] != OS_Unknown {
				orders.hallCalls[floor][1] = remoteOrders.hallCalls[floor][1]
			}
		case OS_None:
			if remoteOrders.hallCalls[floor][1] == OS_Unconfirmed {
				orders.hallCalls[floor][1] = remoteOrders.hallCalls[floor][1]
			}
		case OS_Unconfirmed:
			if len(orders.hallCallConsensus[floor][1]) == num_alive_elevators {
				orders.hallCalls[floor][1] = OS_Confirmed
			}
		}
	}
}

func fsm_aliveElevatorsRecieved(self_id string) {
	alive_counter := 0
	for id, isAlive := range alive_elevators {
		if isAlive {
			alive_counter++
			if _, err := orders.cabCalls[id]; err {
				orders.cabCalls[id] = &[N_FLOORS]OrderState{}
			}
		} else {
			for i := 0; i < N_FLOORS; i++ {
				if orders.cabCalls[id][i] == OS_None {
					orders.cabCalls[id][i] = OS_Unknown
				}
			}
		}
	}
	num_alive_elevators = alive_counter
	if num_alive_elevators == 1 { //DISCONNECTED
		for floor := 0; floor < N_FLOORS; floor++ {
			if orders.cabCalls[self_id][floor] == OS_Unconfirmed {
				orders.cabCalls[self_id][floor] = OS_Confirmed
			}
			if orders.hallCalls[floor][0] == OS_None || orders.hallCalls[floor][0] == OS_Unconfirmed {
				orders.hallCalls[floor][0] = OS_Unknown
			}
			if orders.hallCalls[floor][1] == OS_None || orders.hallCalls[floor][1] == OS_Unconfirmed {
				orders.hallCalls[floor][1] = OS_Unknown
			}
		}
	}
}

func fsm_localOrderServed(served_order elevio.ButtonEvent, id string) {
	floor := served_order.Floor
	btn := served_order.Button
	switch btn {
	case elevio.BT_Cab:
		orders.cabCalls[id][floor] = OS_None
	case elevio.BT_HallUp:
		if num_alive_elevators > 1 {
			orders.hallCalls[floor][0] = OS_None
		} else {
			orders.hallCalls[floor][0] = OS_Unknown
		}
	case elevio.BT_HallDown:
		if num_alive_elevators > 1 {
			orders.hallCalls[floor][1] = OS_None
		} else {
			orders.hallCalls[floor][1] = OS_Unknown
		}
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
