package main

import (
	"Elevator-project/src/alive_listener"
	"Elevator-project/src/elevator_control"
	"Elevator-project/src/elevio"
	"Elevator-project/src/network"
	"Elevator-project/src/network/peers"
	"Elevator-project/src/order_assigner"
	"Elevator-project/src/order_redundancy_manager"
	"flag"
)

const N_FLOORS = 4 //REMOVE THIS
const N_BTN_TYPES = 3
const HARDWARE_ADDR = "localhost:15657"

func main() {

	println("Started!")
	// Get ID of this elevator:
	// `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	//cnahhhels
	// Hardware
	drv_or_buttons := make(chan elevio.ButtonEvent)
	drv_ec_floor := make(chan int)
	drv_ec_obstr := make(chan bool)
	drv_ec_stop := make(chan bool)

	//Elevator control
	ec_or_localOrderServed := make(chan elevio.ButtonEvent)

	//Assigned order
	oa_ec_assignedOrders := make(chan [N_FLOORS][N_BTN_TYPES]bool)
	al_oa_newElevDetected := make(chan string)
	al_oa_elevsLost := make(chan []string)
	// Order redundancy
	or_oa_confirmedOrders := make(chan order_redundancy_manager.ConfirmedOrders)
	net_or_remoteOrders := make(chan order_redundancy_manager.OrdersMSG)
	or_net_localOrders := make(chan order_redundancy_manager.OrdersMSG)

	// //Network
	net_peersUpdate := make(chan peers.PeerUpdate)
	ec_net_elevatorStateTX := make(chan elevator_control.ElevatorState)
	net_oa_elevatorStateRX := make(chan elevator_control.ElevatorState)

	al_or_newElevDetected := make(chan string)
	al_or_elevsLost := make(chan []string)
	al_or_disconnected := make(chan bool)

	elevio.Init(elevator_control.HARDWARE_ADDR, elevator_control.N_FLOORS)
	go alive_listener.AliveListener(net_peersUpdate, al_or_newElevDetected, al_or_elevsLost, al_or_disconnected, al_oa_newElevDetected, al_oa_elevsLost)
	go elevio.PollButtons(drv_or_buttons)
	go elevio.PollFloorSensor(drv_ec_floor)
	go elevio.PollObstructionSwitch(drv_ec_obstr)
	go elevio.PollStopButton(drv_ec_stop)
	go order_assigner.OrderAssigner(or_oa_confirmedOrders, net_oa_elevatorStateRX, al_oa_newElevDetected, al_oa_elevsLost, oa_ec_assignedOrders, id)
	go elevator_control.ElevatorControl(oa_ec_assignedOrders, drv_ec_floor, drv_ec_obstr, drv_ec_stop, ec_net_elevatorStateTX, ec_or_localOrderServed, id)

	go network.Network(id, net_peersUpdate, ec_net_elevatorStateTX, net_oa_elevatorStateRX, or_net_localOrders, net_or_remoteOrders)

	go order_redundancy_manager.OrderRedundancyManager(net_or_remoteOrders,
		drv_or_buttons,
		ec_or_localOrderServed,
		al_or_newElevDetected,
		al_or_elevsLost,
		al_or_disconnected,
		or_oa_confirmedOrders,
		or_net_localOrders,
		id)

	for {
	}
}
