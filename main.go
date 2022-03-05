package main

import (
	"Elevator-project/src/elevator_control"
	"Elevator-project/src/elevio"
	"Elevator-project/src/order_assigner"
	"Elevator-project/src/order_redundancy_manager"
	"Elevator-project/src/order_state_pub"
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
	assigner_assignedOrders := make(chan [N_FLOORS][N_BUTTONS]bool)

	// Order redundancy
	orm_confirmedOrders := make(chan order_redundancy_manager.ConfirmedOrders)
	orm_remoteOrders := make(chan order_redundancy_manager.Orders)
	orm_localOrders := make(chan order_redundancy_manager.Orders)

	// //Network
	// oasTx := make(chan network.OrdersAndStateMessage)
	// oasRx := make(chan network.OrdersAndStateMessage)

	// Order State Publisher
	osp_elevatorState := make(chan map[string]order_state_pub.ElevatorState)

	// Order State Reciever
	osr_alliveElevators := make(chan map[string]bool)

	elevio.Init(elevator_control.HARDWARE_ADDR, elevator_control.N_FLOORS)
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go order_assigner.OrderAssigner(assigner_assignedOrders, orm_confirmedOrders, osp_elevatorState, id)

	go elevator_control.ElevatorControl(assigner_assignedOrders, drv_floors, drv_obstr, drv_stop)

	// go network.Network(id, oasTx, oasRx)

	go order_redundancy_manager.OrderRedundancyManager(orm_confirmedOrders, orm_remoteOrders, osr_alliveElevators, orm_localOrders, drv_buttons, ec_localOrderServed, id)
	for {
	}
}
