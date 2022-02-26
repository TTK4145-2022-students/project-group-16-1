package main

import (
	"Elevator-project/src/elevator_control"
	"Elevator-project/src/elevio"
	"Elevator-project/src/network"
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

	// Hardware channels
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	//Assigned order
	assigner_assignedOrders := make(chan [N_FLOORS][N_BUTTONS]bool)

	// Order redundancy channels
	orm_confirmedOrders := make(chan order_redundancy_manager.ConfirmedOrders)
	orm_remoteOrders := make(chan order_redundancy_manager.OrdersAndAliveMSG)
	orm_localOrders := make(chan order_redundancy_manager.Orders)

	//Network channels
	oasTx := make(chan network.OrdersAndStateMessage)
	oasRx := make(chan network.OrdersAndStateMessage)

	elevio.Init(elevator_control.HARDWARE_ADDR, elevator_control.N_FLOORS)
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go elevator_control.ElevatorControl(assigner_assignedOrders, drv_floors, drv_obstr, drv_stop)

	go network.Network(id, oasTx, oasRx)

	go order_redundancy_manager.OrderRedundancyManager(orm_confirmedOrders, orm_remoteOrders, orm_localOrders)
	for {
	}
}
