package main

import (
	"Elevator-project/src/elevator_control"
	"Elevator-project/src/elevio"
)


func main() {
	println("Started!")

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	elevio.Init(elevator_control.HARDWARE_ADDR, elevator_control.N_FLOORS)
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go elevator_control.Elevator_control(drv_buttons, drv_floors, drv_obstr, drv_stop)

	for {
	}
}
