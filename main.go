package main

import (
	"Elevator-project/src/alive_listener"
	. "Elevator-project/src/constants"
	"Elevator-project/src/elevator_control"
	"Elevator-project/src/elevator_io"
	"Elevator-project/src/network"
	"Elevator-project/src/network/peers"
	"Elevator-project/src/order_assigner"
	"Elevator-project/src/order_redundancy"
	"flag"
)

func main() {
	// Get ID of this elevator:
	// `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// Abbreviations:
	// eio: elevator io driver
	// ec: elevator_control - control of physical elevator. Executes orders. 
	// oa: order assigner
	// al: alive listener - keeps track of elevators on network
	// or: order_redundancy - state machine/cyclic counter for backing up orders. Peer to peer.
	// net: network

	// Channels:
	// Naming convention: from_to_name

	eio_or_buttons 				:= make(chan elevator_io.ButtonEvent, 1)
	eio_ec_floor 				:= make(chan int, 1)
	eio_ec_obstr 				:= make(chan bool, 1)

	ec_or_localOrderServed 		:= make(chan elevator_io.ButtonEvent, 1)
	ec_oa_elevatorState 		:= make(chan elevator_control.ElevatorStateMsg, 1)
	ec_net_elevatorState 		:= make(chan elevator_control.ElevatorStateMsg, 1)

	oa_ec_localAssignedOrders 	:= make(chan [N_FLOORS][N_BTN_TYPES]bool, 1)

	al_oa_newElevDetected 		:= make(chan string, 1)
	al_or_newElevDetected 		:= make(chan string, 1)
	al_oa_elevsLost 			:= make(chan []string, 1)
	al_or_elevsLost 			:= make(chan []string, 1)
	al_or_disconnected 			:= make(chan bool, 1)

	or_oa_confirmedOrders 		:= make(chan order_redundancy.ConfirmedOrders, 1)
	or_net_localOrders 			:= make(chan order_redundancy.OrdersMSG, 1)

	net_al_peersUpdate 			:= make(chan peers.PeerUpdate, 1)
	net_or_remoteOrders 		:= make(chan order_redundancy.OrdersMSG, 1)
	net_oa_elevatorState 		:= make(chan elevator_control.ElevatorStateMsg, 1)

	// Initialising hardware:
	elevator_io.Init(HARDWARE_ADDR, N_FLOORS)

	// Modules: 
	go alive_listener.AliveListener(
		// Input
		net_al_peersUpdate,
		// Output
		al_or_newElevDetected,
		al_or_elevsLost,
		al_or_disconnected,
		al_oa_newElevDetected,
		al_oa_elevsLost,
		// ID
		id)
	go elevator_io.PollButtons(
		// Output
		eio_or_buttons,
	)
	go elevator_io.PollFloorSensor(
		// Output
		eio_ec_floor,
	)
	go elevator_io.PollObstructionSwitch(
		// Output
		eio_ec_obstr,
	)
	go order_assigner.OrderAssigner(
		// Input
		or_oa_confirmedOrders,
		net_oa_elevatorState,
		ec_oa_elevatorState,
		al_oa_newElevDetected,
		al_oa_elevsLost,
		// Output
		oa_ec_localAssignedOrders,
		// ID
		id)
	go elevator_control.ElevatorControl(
		// Input
		oa_ec_localAssignedOrders,
		eio_ec_floor,
		eio_ec_obstr,
		// Output
		ec_net_elevatorState,
		ec_oa_elevatorState,
		ec_or_localOrderServed,
		// ID
		id)
	go network.Network(
		// Input
		ec_net_elevatorState,
		or_net_localOrders,
		// Ouput
		net_oa_elevatorState,
		net_or_remoteOrders,
		net_al_peersUpdate,
		// ID
		id)
	go order_redundancy.OrderRedundancy(
		// Input
		net_or_remoteOrders,
		eio_or_buttons,
		ec_or_localOrderServed,
		al_or_newElevDetected,
		al_or_elevsLost,
		al_or_disconnected,
		// Output
		or_oa_confirmedOrders,
		or_net_localOrders,
		// ID
		id)

	select {}
}
