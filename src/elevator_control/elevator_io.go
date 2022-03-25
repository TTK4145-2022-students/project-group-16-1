package elevator_control

// Mapping from elevio module to elevator control module

import (
	"Elevator-project/src/elevio"
	"fmt"
)

type Dirn int

const (
	D_Down Dirn = -1
	D_Stop Dirn = 0
	D_Up   Dirn = 1
)

func (dirn Dirn) String() string {
	dirnStrings := [...]string{"down", "stop", "up"}
	if dirn < D_Down || dirn > D_Up {
		return fmt.Sprintf("A non declared behaviour was given: (%d)", int(dirn))
	}
	return dirnStrings[dirn+1]
}

type Button int

const (
	B_HallUp   Button = 0
	B_HallDown Button = 1
	B_Cab      Button = 2
)

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

func createElevatorStateMSG(elevator *Elevator, id string) ElevatorState {
	var elevator_state ElevatorState
	elevator_state.Behaviour = elevator.behaviour.String()
	elevator_state.Dirn = elevator.dirn.String()
	elevator_state.Floor = elevator.floor
	elevator_state.Id = id
	elevator_state.Available = !((elevator.obstructed && elevator.behaviour == EB_DoorOpen) || elevator.too_late)
	return elevator_state
}
