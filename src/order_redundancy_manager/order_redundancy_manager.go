package order_redundancy_manager

import (
	"fmt"
	"time"
)

const N_FLOORS = 4 //REMOVE THIS
const N_BUTTONS = 3
const HARDWARE_ADDR = "localhost:15657"

var aliveElevators []string
var orders Orders

type OrderState int

const (
	Unknown OrderState = iota
	Confirmed
	None
	Unconfirmed
)

type Orders struct {
	hallCalls [N_FLOORS][2]OrderState
	cabCalls  [N_FLOORS][]OrderState
}

type OrdersAndAliveMSG struct {
	orders         Orders
	aliveElevators []string
}

type ConfirmedOrders struct {
	HallCalls [N_FLOORS][2]bool
	CabCalls  map[string][N_FLOORS]bool
}

func OrderRedundancyManager(orm_confirmedOrders chan ConfirmedOrders, orm_ordersAndAlive chan OrdersAndAliveMSG, orm_localOrders chan Orders) {
	// for {
	// 	select {
	// 	case ordersAndAlive := <-orm_ordersAndAlive:
	// 		aliveElevators = ordersAndAlive.aliveElevators
	// 	}
	// }
	time.Sleep(7 * time.Second)
	var orders ConfirmedOrders

	orders.HallCalls = [N_FLOORS][2]bool{{false, false}, {true, false}, {false, false}, {false, false}}
	orders.CabCalls = make(map[string][N_FLOORS]bool)
	orders.CabCalls["one"] = [N_FLOORS]bool{false, false, true, true}
	orders.CabCalls["two"] = [N_FLOORS]bool{false, false, false, false}
	fmt.Println("About to send orders")
	orm_confirmedOrders <- orders
	fmt.Println("Have sent orders")
}
