package main

import (
	"Elevator-project/src/elevator_control"
	"Elevator-project/src/elevio"
)

const N_FLOORS = 4 //REMOVE THIS
const N_BUTTONS = 3
const HARDWARE_ADDR = "localhost:15657"

func main() {
	println("Started!")

	// Hardware channels
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	//Assigned order
	assigner_assignedOrders := make(chan [N_FLOORS][N_BUTTONS]bool)

	elevio.Init(elevator_control.HARDWARE_ADDR, elevator_control.N_FLOORS)
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go elevator_control.Elevator_control(assigner_assignedOrders, drv_floors, drv_obstr, drv_stop)

	for {
	}
}
