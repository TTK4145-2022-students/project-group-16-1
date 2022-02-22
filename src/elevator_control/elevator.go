package elevator_control

import (
	"fmt"
	"time"
)

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type ClearRequestVariant int

const (
	CV_All ClearRequestVariant = iota
	CV_InDirn
)

type Elevator struct {
	floor     int
	dirn      Dirn
	requests  [N_FLOORS][N_BUTTONS]int
	behaviour ElevatorBehaviour
	config    config
}

type config struct {
	clearRequestVariant ClearRequestVariant
	doorOpenDuration_s  time.Duration
}

func eb_toString(ev ElevatorBehaviour) {
	fmt.Println("Print elev behaviour here")
}

func elevator_print(elevator Elevator) {
	fmt.Println("Print elev here")
}

func elevator_uninitialised() Elevator {
	return Elevator{
		floor:     -1,
		dirn:      D_Stop,
		behaviour: EB_Idle,
		config: config{
			clearRequestVariant: CV_InDirn,
			doorOpenDuration_s:  3 * time.Second},
	}
}
