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

func (dirn Dirn) String() string {
	dirnStrings := [...]string{"D_Down", "D_Stop", "D_Up"}
	if dirn < D_Down || dirn > D_Up {
		return fmt.Sprintf("A non declared behaviour was given: (%d)", int(dirn))
	}
	return dirnStrings[dirn+1]
}

type Button int

const (
	B_HallUp   Button = 0
	B_HallDown        = 1
	B_Cab             = 2
)

func initElevIO() {
	elevio.Init(HARDWARE_ADDR, N_FLOORS)
}

func io_setMotorDirection(dir Dirn) {
	elevio.SetMotorDirection(elevio.MotorDirection(dir))
}

func io_setButtonLamp(button Button, floor int, value bool) {
	elevio.SetButtonLamp(elevio.ButtonType(button), floor, value)
}

func io_setFloorIndicator(floor int) {
	elevio.SetFloorIndicator(floor)
}

func io_setDoorOpenLamp(value bool) {
	elevio.SetDoorOpenLamp(value)
}

func _setStopLamp(value bool) {
	elevio.SetStopLamp(value)
}
