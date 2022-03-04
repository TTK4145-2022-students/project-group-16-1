package network

import (
	"Elevator-project/src/network/bcast"
	"Elevator-project/src/network/peers"
	"fmt"
)

type OrdersAndStateMessage struct {
	Id int
	// State  Elevator
	// Orders Orders
}

func Network(id string, oasTX chan OrdersAndStateMessage, oasRX chan OrdersAndStateMessage) {

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(27773, id, peerTxEnable)
	go peers.Receiver(27773, peerUpdateCh)

	go bcast.Transmitter(27772, oasTX)
	go bcast.Receiver(27772, oasRX)

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
