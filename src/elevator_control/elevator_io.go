package elevator_control

// Mapping from elevio module to elevator control module

import (
	"Elevator-project/src/elevio"
	"fmt"
)

type Dirn int

const (
	D_Down Dirn = -1
	D_Stop      = 0
	D_Up        = 1
)

type Button int

const (
	B_HallUp   Button = 0
	B_HallDown        = 1
	B_Cab             = 2
)

func initElevIO() {
	elevio.Init()
}

func setMotorDirection(dir Dirn) {
	elevio.SetMotorDirection(dir)
}

func setButtonLamp(button Button, floor int, value bool) {
	elevio.SetButtonLamp(button,floor, value)
}

func setFloorIndicator(floor int) {
	elevio.setFloorIndicator(floor)
}

func setDoorOpenLamp(value bool) {
	elevio.SetDoorOpenLamp(value)
}

func setStopLamp(value bool) {
	elevio.SetStopLamp(value)
}

