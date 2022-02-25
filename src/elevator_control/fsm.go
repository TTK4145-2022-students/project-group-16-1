package elevator_control

import (
	"fmt"
	"time"
)

var elevator *Elevator
var door_timer *time.Timer

const N_FLOORS = 4 //REMOVE THIS
const N_BUTTONS = 3
const HARDWARE_ADDR = "localhost:15657"

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
		for btn := 0; btn < N_BUTTONS; btn++ {
			io_setButtonLamp(Button(btn), floor, elevator.requests[floor][btn] != 0)
		}
	}
}

func fsm_onInitBetweenFloors() {
	io_setMotorDirection(D_Down)
	elevator.dirn = D_Down
	elevator.behaviour = EB_Moving
}

func fsm_onRequestButtonPress(btn_floor int, btn_type Button) {
	fmt.Println("--------------")
	fmt.Println("Jumping into [fsm_onRequestButtonPress]")
	elevator_print(elevator)

	switch elevator.behaviour {
	case EB_DoorOpen, EB_Obstructed:
		if requests_shouldClearImmediately(elevator, btn_floor, btn_type) {
			door_timer = time.NewTimer(elevator.config.doorOpenDuration_s)
		} else {
			elevator.requests[btn_floor][btn_type] = 1
		}

	case EB_Moving:
		elevator.requests[btn_floor][btn_type] = 1

	case EB_Idle:
		elevator.requests[btn_floor][btn_type] = 1
		a := requests_nextAction(elevator)

		elevator.dirn = a.dirn
		elevator.behaviour = a.behaviour
		switch elevator.behaviour {
		case EB_DoorOpen:
			io_setDoorOpenLamp(true)
			door_timer = time.NewTimer(elevator.config.doorOpenDuration_s)
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
			door_timer = time.NewTimer(elevator.config.doorOpenDuration_s)
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
			door_timer = time.NewTimer(elevator.config.doorOpenDuration_s)
			requests_clearAtCurrentFloor(elevator)
			setAllLights(elevator)
		case EB_Idle, EB_Moving:
			io_setDoorOpenLamp(false)
			io_setMotorDirection(elevator.dirn)
		}
	case EB_Obstructed:
		door_timer = time.NewTimer(elevator.config.doorOpenDuration_s)
	default:
	}
	fmt.Println("New state:")
	elevator_print(elevator)
	fmt.Println("--------------")
}

func fsm_onObstruction(obstructed bool) {
	fmt.Println("--------------")
	fmt.Println("Jumping into [fsm_onObstruction")
	elevator_print(elevator)
	switch elevator.behaviour {
	case EB_DoorOpen:
		if obstructed {
			elevator.behaviour = EB_Obstructed
		}
	case EB_Obstructed:
		if !obstructed {
			elevator.behaviour = EB_DoorOpen
		}
	default:
	}
	fmt.Println("New state:")
	elevator_print(elevator)
	fmt.Println("--------------")

}
