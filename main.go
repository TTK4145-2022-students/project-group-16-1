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

	//Channels:
	//Naming convention: from_to_name

	//eio: elevator io driver
	eio_or_buttons 			:= make(chan elevator_io.ButtonEvent, 1)
	eio_ec_floor 			:= make(chan int, 1)
	eio_ec_obstr 			:= make(chan bool, 1)

	//ec: elevator_control - control of physical elevator
	ec_or_localOrderServed 	:= make(chan elevator_io.ButtonEvent, 1)
	ec_oa_elevatorState 	:= make(chan elevator_control.ElevatorStateMsg, 1)
	ec_net_elevatorStateTX 	:= make(chan elevator_control.ElevatorStateMsg, 1)

	//oa: order assigner
	oa_ec_assignedOrders 	:= make(chan [N_FLOORS][N_BTN_TYPES]bool, 1)

	//al: alive listener - keeps track of elevators on network
	al_oa_newElevDetected 	:= make(chan string, 1)
	al_or_newElevDetected 	:= make(chan string, 1)
	al_oa_elevsLost 		:= make(chan []string, 1)
	al_or_elevsLost 		:= make(chan []string, 1)
	al_or_disconnected 		:= make(chan bool, 1)

	//or:order_redundancy - state machine/cyclic counter for orders
	or_oa_confirmedOrders 	:= make(chan order_redundancy.ConfirmedOrders, 1)
	or_net_localOrders 		:= make(chan order_redundancy.OrdersMSG, 1)

	//net: network
	net_al_peersUpdate 		:= make(chan peers.PeerUpdate, 1)
	net_or_remoteOrders 	:= make(chan order_redundancy.OrdersMSG, 1)
	net_oa_elevatorStateRX 	:= make(chan elevator_control.ElevatorStateMsg, 1)

	elevator_io.Init(HARDWARE_ADDR, N_FLOORS)

	go alive_listener.AliveListener(
		net_al_peersUpdate,
		al_or_newElevDetected,
		al_or_elevsLost,
		al_or_disconnected,
		al_oa_newElevDetected,
		al_oa_elevsLost,
		id)
	go elevator_io.PollButtons(eio_or_buttons)
	go elevator_io.PollFloorSensor(eio_ec_floor)
	go elevator_io.PollObstructionSwitch(eio_ec_obstr)
	go order_assigner.OrderAssigner(
		or_oa_confirmedOrders,
		net_oa_elevatorStateRX,
		ec_oa_elevatorState,
		al_oa_newElevDetected,
		al_oa_elevsLost,
		oa_ec_assignedOrders,
		id)
	go elevator_control.ElevatorControl(
		oa_ec_assignedOrders,
		eio_ec_floor,
		eio_ec_obstr,
		ec_net_elevatorStateTX,
		ec_oa_elevatorState,
		ec_or_localOrderServed,
		id)
	go network.Network(
		net_al_peersUpdate,
		ec_net_elevatorStateTX,
		net_oa_elevatorStateRX,
		or_net_localOrders,
		net_or_remoteOrders,
		id)
	go order_redundancy.OrderRedundancy(
		net_or_remoteOrders,
		eio_or_buttons,
		ec_or_localOrderServed,
		al_or_newElevDetected,
		al_or_elevsLost,
		al_or_disconnected,
		or_oa_confirmedOrders,
		or_net_localOrders,
		id)

	select {}
}
