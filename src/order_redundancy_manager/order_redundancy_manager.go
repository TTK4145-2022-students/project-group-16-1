package order_redundancy_manager

import (
	"Elevator-project/src/elevio"
)

const N_FLOORS = 4 //REMOVE THIS
const N_BUTTONS = 3
const HARDWARE_ADDR = "localhost:15657"

var aliveElevators map[string]bool
var orders Orders
var isConnected bool

type OrderState int

const (
	OS_Unknown OrderState = iota
	OS_Confirmed
	OS_None
	OS_Unconfirmed
)

type Orders struct {
	hallCalls [N_FLOORS][2]OrderState
	cabCalls  map[string]*[N_FLOORS]OrderState
}

type ConfirmedOrders struct {
	HallCalls [N_FLOORS][2]bool
	CabCalls  map[string][N_FLOORS]bool
}

func OrderRedundancyManager(orm_confirmedOrders chan ConfirmedOrders,
	orm_remoteOrders chan Orders,
	osr_alliveElevators chan map[string]bool,
	orm_localOrders chan Orders, drv_buttons chan elevio.ButtonEvent,
	ec_localOrderServed chan elevio.ButtonEvent, id string) {

	isConnected = false
	aliveElevators = make(map[string]bool)
	aliveElevators[id] = true

	orders.cabCalls = make(map[string]*[N_FLOORS]OrderState)
	orders.cabCalls[id] = &[N_FLOORS]OrderState{}
	for {
		select {
		case aliveElevators = <-osr_alliveElevators:
			fsm_aliveElevatorsRecieved()
		case remoteOrders := <-orm_remoteOrders:
			fsm_remoteOrdersRecieved(remoteOrders)
		case button_event := <-drv_buttons:
			fsm_buttonPres = &[N_FLOORS]OrderState{}

	// time.Sleep(7 * time.Second)
	// var orders ConfirmedOrders

	// orders.HallCalls = [N_FLOORS][2]bool{{false, false}, {true, false}, {false, false}, {false, false}}
	// orders.CabCalls = make(map[string][N_FLOORS]bool)
	// orders.CabCalls["one"] = [N_FLOORS]bool{false, false, true, true}
	// orders.CabCalls["two"] = [N_FLOORS]bool{false, false, false, false}
	// fmt.Println("About to send orders")
	// orm_confirmedOrders <- orders
	// fmt.Println("Have sent orders")
}}}
func fsm_buttonPressed(button_event elevio.ButtonEvent, id string) {
	floor := button_event.Floor
	switch button_event.Button {
	case elevio.BT_Cab:
		switch orders.cabCalls[id][floor] {
		case OS_Unknown, OS_None:
			orders.cabCalls[id][floor] = OS_Unconfirmed
		}
	case elevio.BT_HallUp:
		if isConnected {
			switch orders.hallCalls[floor][0] {
			case OS_Unknown, OS_None:
				orders.hallCalls[floor][0] = OS_Unconfirmed
			}
		}
	case elevio.BT_HallDown:
		if isConnected {
			switch orders.hallCalls[floor][1] {
			case OS_Unknown, OS_None:
				orders.hallCalls[floor][1] = OS_Unconfirmed
			}
		}
	}
}

func fsm_remoteOrdersRecieved(remoteOrders Orders) {
	for id, calls := range remoteOrders.cabCalls {
		for floor := 0; floor < N_FLOORS; floor++ {
			switch orders.cabCalls[id][floor]{
			case OS_Unknown:
				orders.cabCalls[id][floor] = remoteOrders.cabCalls[id][floor]
			case OS_Confirmed:
				if remoteOrders.cabCalls[id][floor] != OS_Unknown{
					orders.cabCalls[id][floor] = remoteOrders.cabCalls[id][floor]
				}
			case OS_None:
				if remoteOrders.cabCalls[id][floor] == OS_Unconfirmed{
					orders.cabCalls[id][floor] = remoteOrders.cabCalls[id][floor]
				}
			}
		}
	}

	for floor:=0;floor<N_FLOORS;floor++{
		switch orders.HallCalls[0][floor]
	}
}

func fsm_aliveElevatorsRecieved(self_id string) {
	allive_elevs := 0
	for id, isAlive := range aliveElevators {
		if isAlive {
			allive_elevs++
			if val, err := orders.cabCalls[id]; err {
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
	if allive_elevs == 1 {
		isConnected = false
		for i := 0; i < N_FLOORS; i++ {
			if orders.cabCalls[self_id][i] == OS_Unconfirmed {
				orders.cabCalls[self_id][i] = OS_Confirmed
			}
			if orders.HallCalls[0] == OS_None || orders.HallCalls[0] == OS_Unconfirmed {
				orders.HallCalls[0] == OS_Unknown
			}
			if orders.HallCalls[1] == OS_None || orders.HallCalls[1] == OS_Unconfirmed {
				orders.HallCalls[1] == OS_Unknown
			}
		}
	} else {
		isConnected = true
	}
}
