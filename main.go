package main

import (
	"Elevator-project/src/elevator_control"
	"Elevator-project/src/elevio"
)

func main() {
	println("Started!")

	numFloors := 4
	inputPollRate_ms := 0.25

	elevio.Init("localhost:15657", numFloors)

	var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go elevator_control.Elevator_control(numFloors, inputPollRate_ms, drv_buttons, drv_floors, drv_obstr, drv_stop)
}
