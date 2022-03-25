package elevator_control

import (
	"fmt"
)

func elevator_init(elevator *Elevator) {
	fmt.Println("Initialising FSM: ")
	*elevator = elevator_uninitialised()
	if elevator.floor == -1 {
		elevator_onInitBetweenFloors(elevator)
	}

}

func elevator_onInitBetweenFloors(elevator *Elevator) {
	io_setMotorDirection(D_Down)
	elevator.dirn = D_Down
	elevator.behaviour = EB_Moving

}
