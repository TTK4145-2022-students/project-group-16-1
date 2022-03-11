package order_state_pub

import (
	"Elevator-project/src/elevator_control"
	"Elevator-project/src/order_redundancy_manager"
)

type OrdersAndStateMSG struct {
	Id     string
	State  elevator_control.ElevatorState
	Orders order_redundancy_manager.Orders
}

func order_state_pub() {

}
