package elevator_control

import (
	"fmt"
)

func fsm_init() {
	fmt.Println("Initialising FSM: ")
	elevator = elevator_uninitialised()
	if elevator.floor == -1 {
		fsm_onInitBetweenFloors()
	}
	elevator_print(elevator)
}

func setAllLights(es *Elevator) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BTN_TYPES; btn++ {
			io_setButtonLamp(Button(btn), floor, elevator.requests[floor][btn])
		}
	}
}

func fsm_onInitBetweenFloors() {
	io_setMotorDirection(D_Down)
	elevator.dirn = D_Down
	elevator.behaviour = EB_Moving
}

func fsm_onRequestUpdate(a [N_FLOORS][N_BTN_TYPES]bool) {
	fmt.Println("--------------")
	fmt.Println("Jumping into [fsm_onRequestUpdate]")
	elevator_print(elevator)

	elevator.requests = a

	switch elevator.behaviour {
	case EB_DoorOpen, EB_Obstructed:
		requests_shouldClearImmediately(elevator)

	case EB_Idle:
		a := requests_nextAction(elevator)
		elevator.dirn = a.dirn
		elevator.behaviour = a.behaviour
		switch elevator.behaviour {
		case EB_DoorOpen:
			io_setDoorOpenLamp(true)
			door_timer.Reset(elevator.config.doorOpenDuration_s)
			requests_clearAtCurrentFloor(elevator)
		case EB_Moving:
			io_setMotorDirection(elevator.dirn)
		case EB_Idle:

		}

	}
	setAllLights(elevator)

	fmt.Println("New state:")
	elevator_print(elevator)
	fmt.Println("--------------")
}

func fsm_onFloorArrival(newFloor int) {
	fmt.Println("--------------")
	fmt.Println("Jumping into [fsm_onFloorArrival]")
	elevator_print(elevator)

	elevator.floor = newFloor

	io_setFloorIndicator(newFloor)

	switch elevator.behaviour {
	case EB_Moving:
		if requests_shouldStop(elevator) {
			io_setMotorDirection(D_Stop)
			io_setDoorOpenLamp(true)
			requests_clearAtCurrentFloor(elevator)
			door_timer.Reset(elevator.config.doorOpenDuration_s)
			setAllLights(elevator)
			elevator.behaviour = EB_DoorOpen
		}
	default:
	}

	fmt.Println("New state:")
	elevator_print(elevator)
	fmt.Println("--------------")
}

func fsm_onDoorTimeout() {
	fmt.Println("Jumping into [fsm_onDoorTimeout]")
	elevator_print(elevator)
	switch elevator.behaviour {
	case EB_DoorOpen:
		a := requests_nextAction(elevator)
		elevator.dirn = a.dirn
		elevator.behaviour = a.behaviour

		switch elevator.behaviour {
		case EB_DoorOpen:
			door_timer.Reset(elevator.config.doorOpenDuration_s)
			requests_clearAtCurrentFloor(elevator)
			setAllLights(elevator)
		case EB_Idle, EB_Moving:
			io_setDoorOpenLamp(false)
			io_setMotorDirection(elevator.dirn)
		}
	case EB_Obstructed:
		door_timer.Reset(elevator.config.doorOpenDuration_s)
	default:
	}
	fmt.Println("New state:")
	elevator_print(elevator)
	fmt.Println("--------------")
}
