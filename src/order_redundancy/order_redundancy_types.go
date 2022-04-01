package order_redundancy

import (
	. "Elevator-project/src/constants"
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
	CabCalls  map[string][N_FLOORS]bool
}

func createOrdersMSG(id string, orders Orders) OrdersMSG {
	var orders_msg OrdersMSG
	orders_msg.Id 					= id
	orders_msg.CabCalls 			= make(map[string][N_FLOORS]OrderState)
	orders_msg.HallCalls 			= orders.HallCalls
	orders_msg.HallCallConsensus 	= orders.HallCallConsensus
	orders_msg.CabCallConsensus 	= make(map[string][N_FLOORS][]string)
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
	confirmed_orders.CabCalls = make(map[string][N_FLOORS]bool)

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++ {
			if orders.HallCalls[floor][btn] == OS_Confirmed {
				confirmed_orders.HallCalls[floor][btn] = true
			}
		}
	}
	for _, elev_id := range alive_elevators {
		confirmed_orders.CabCalls[elev_id] = [N_FLOORS]bool{}
		cab_calls := [N_FLOORS]bool{}
		for floor := 0; floor < N_FLOORS; floor++ {
			if orders.CabCalls[elev_id][floor] == OS_Confirmed {
				cab_calls[floor] = true
			}
		}
		confirmed_orders.CabCalls[elev_id] = cab_calls
	}
	confirmed_orders.CabCalls[id] = [N_FLOORS]bool{}
	cab_calls := [N_FLOORS]bool{}
	for floor := 0; floor < N_FLOORS; floor++ {
		if orders.CabCalls[id][floor] == OS_Confirmed {
			cab_calls[floor] = true
		}
	}
	confirmed_orders.CabCalls[id] = cab_calls
	return confirmed_orders
}
