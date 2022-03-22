package elevator_control

import (
	"fmt"
)

func fsm_init(elevator *Elevator) {
	fmt.Println("Initialising FSM: ")
	*elevator = elevator_uninitialised()
	if elevator.floor == -1 {
		fsm_onInitBetweenFloors(elevator)
	}

}

func fsm_onInitBetweenFloors(elevator *Elevator) {
	io_setMotorDirection(D_Down)
	elevator.dirn = D_Down
	elevator.behaviour = EB_Moving

}
