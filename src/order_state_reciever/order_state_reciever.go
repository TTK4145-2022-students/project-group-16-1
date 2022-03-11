package order_state_reciever

import (
	"Elevator-project/src/elevator_control"
	"Elevator-project/src/network/peers"
	"Elevator-project/src/order_redundancy_manager"
	"Elevator-project/src/order_state_pub"
)

func OrderStateReciever(
	orderStateRx <-chan order_state_pub.OrdersAndStateMSG,
	remoteOrders chan<- order_redundancy_manager.Orders,

) {
	states := make(map[string]elevator_control.ElevatorState)
	for {
		select {
		case orderAndState := <-orderStateRx:
			om orderAndState.id er i state:
				remoteOrders <- orderAndState.Orders
				if orderAndState.State != states[orderAndState.Id] {

				}
		}

	}
}
