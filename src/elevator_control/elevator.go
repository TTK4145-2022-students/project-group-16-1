package elevator_control

import (
	. "Elevator-project/src/constants"
	"fmt"
)

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type Elevator struct {
	floor      int
	dirn       Dirn
	requests   [N_FLOORS][N_BTN_TYPES]bool
	behaviour  ElevatorBehaviour
	obstructed bool
	too_late   bool
}

type ElevatorState struct {
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
		dirn:      D_Stop,
		behaviour: EB_Idle,
	}

	return elevator
}
