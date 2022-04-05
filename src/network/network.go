package network

import (
	"Elevator-project/src/elevator_control"
	"Elevator-project/src/network/bcast"
	"Elevator-project/src/network/peers"
	"Elevator-project/src/order_redundancy"
)

func Network(
	stateTX 		chan elevator_control.ElevatorStateMsg,
	orderTX 		chan order_redundancy.OrdersMSG,
	stateRX 		chan elevator_control.ElevatorStateMsg,
	orderRX 		chan order_redundancy.OrdersMSG,
	peerUpdateCh 	chan peers.PeerUpdate,
	id 				string,
	) {

	peerTxEnable := make(chan bool)
	go peers.Transmitter(27773, id, peerTxEnable)
	go peers.Receiver(27773, peerUpdateCh)

	go bcast.Transmitter(27772, stateTX)
	go bcast.Receiver(27772, stateRX)

	go bcast.Transmitter(27771, orderTX)
	go bcast.Receiver(27771, orderRX)

}
