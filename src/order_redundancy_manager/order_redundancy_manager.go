package order_redundancy_manager

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
	hallCalls [N_FLOORS][2]bool
	cabCalls  [N_FLOORS][]bool
}

func OrderRedundancyManager(orm_confirmedOrders chan ConfirmedOrders, orm_ordersAndAlive chan OrdersAndAliveMSG, orm_localOrders chan Orders) {
	for {
		select {
		case ordersAndAlive := <-orm_ordersAndAlive:
			aliveElevators = ordersAndAlive.aliveElevators
		}
	}
}
