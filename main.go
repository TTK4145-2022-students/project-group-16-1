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
const N_BUTTONS = 3
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
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	//Elevator control
	ec_localOrderServed := make(chan elevio.ButtonEvent)

	//Assigned order
	oa_assignedOrders := make(chan [N_FLOORS][N_BUTTONS]bool)
	oa_newElevDetected := make(chan string)
	oa_elevsLost := make(chan []string)
	// Order redundancy
	orm_confirmedOrders := make(chan order_redundancy_manager.ConfirmedOrders)
	orm_remoteOrders := make(chan order_redundancy_manager.Orders)
	orm_localOrders := make(chan order_redundancy_manager.Orders)

	// //Network
	net_peersUpdate := make(chan peers.PeerUpdate)
	net_elevatorStateTX := make(chan elevator_control.ElevatorState)
	net_elevatorStateRX := make(chan elevator_control.ElevatorState)

	orm_newElevDetected := make(chan string)
	orm_elevsLost := make(chan []string)
	orm_disconnected := make(chan bool)

	elevio.Init(elevator_control.HARDWARE_ADDR, elevator_control.N_FLOORS)
	go alive_listener.AliveListener(net_peersUpdate, orm_newElevDetected, orm_elevsLost, orm_disconnected, oa_newElevDetected, oa_elevsLost)
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go order_assigner.OrderAssigner(orm_confirmedOrders, net_elevatorStateRX, oa_newElevDetected, oa_elevsLost, oa_assignedOrders, id)
	go elevator_control.ElevatorControl(oa_assignedOrders, drv_floors, drv_obstr, drv_stop, net_elevatorStateTX, ec_localOrderServed)

	go network.Network(id, net_peersUpdate, net_elevatorStateTX, net_elevatorStateRX, orm_localOrders, orm_remoteOrders)

	go order_redundancy_manager.OrderRedundancyManager(orm_remoteOrders,
		drv_buttons,
		ec_localOrderServed,
		orm_newElevDetected,
		orm_elevsLost,
		orm_disconnected,
		orm_confirmedOrders,
		orm_localOrders,
		id)

	for {
	}
}
