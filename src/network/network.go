package network

import (
	"Elevator-project/src/elevator_control"
	"Elevator-project/src/network/bcast"
	"Elevator-project/src/network/peers"
	"Elevator-project/src/order_redundancy_manager"
	"fmt"
)

func Network(id string,
	peerUpdateCh chan peers.PeerUpdate,
	stateTX chan elevator_control.ElevatorState,
	stateRX chan elevator_control.ElevatorState,
	orderTX chan order_redundancy_manager.OrdersMSG,
	orderRX chan order_redundancy_manager.OrdersMSG) {

	peerTxEnable := make(chan bool)
	go peers.Transmitter(27773, id, peerTxEnable)
	go peers.Receiver(27773, peerUpdateCh)

	go bcast.Transmitter(27772, stateTX)
	go bcast.Receiver(27772, stateRX)

	go bcast.Transmitter(27772, orderTX)
	go bcast.Receiver(27772, orderRX)

	fmt.Println("Network module started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
		}
	}
}
