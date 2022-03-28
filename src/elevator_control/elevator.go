package elevator_control

import (
	. "Elevator-project/src/constants"
	"Elevator-project/src/elevio"
	"fmt"
)

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type Elevator struct {
	floor         int
	dirn          elevio.MotorDirection
	requests      [N_FLOORS][N_BTN_TYPES]bool
	behaviour     ElevatorBehaviour
	obstructed    bool
	motor_failure bool
}

type ElevatorMsg struct {
	Behaviour string
	Floor     int
	Dirn      string
	Id        string
	Available bool
}

func (ev ElevatorBehaviour) String() string {
	behaviours := [...]string{"idle", "doorOpen", "moving"}
	if ev < EB_Idle || ev > EB_Moving {
		return fmt.Sprintf("An undefined behaviour was given: (%d)", int(ev))
	}
	return behaviours[ev]
}

func elevator_print(elevator *Elevator) {
	fmt.Printf("Floor: %d\n Direction: %s\n Requests: %v\n Behaviour: %s\n", elevator.floor, elevator.dirn.String(), elevator.requests, elevator.behaviour.String())
}

func elevator_uninitialised() Elevator {
	elevator := Elevator{
		floor:     -1,
		dirn:      elevio.MD_Stop,
		behaviour: EB_Idle,
	}

	return elevator
}

func createElevatorMsg(elevator *Elevator, id string) ElevatorMsg {
	var elevator_state ElevatorMsg
	elevator_state.Behaviour = elevator.behaviour.String()
	elevator_state.Dirn = elevator.dirn.String()
	elevator_state.Floor = elevator.floor
	elevator_state.Id = id
	elevator_state.Available = !((elevator.obstructed && elevator.behaviour == EB_DoorOpen) || elevator.motor_failure)
	return elevator_state
}
