package elevator_control

import (
	"fmt"
	"time"
)

var elevator Elevator
var door_timer *time.Timer

const N_FLOORS = 4 //REMOVE THIS
const N_BUTTONS = 3
const HARDWARE_ADDR = ":15657"

func fsm_init() {
	elevator = elevator_uninitialised()
}

func setAllLights(es Elevator) {
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
	//Println("\n\n%s(%d, %s)\n", __FUNCTION__, btn_floor, elevio_button_toString(btn_type))
	elevator_print(elevator)

	switch elevator.behaviour {
	case EB_DoorOpen:
		if requests_shouldClearImmediately(elevator, btn_floor, btn_type) {
			door_timer = time.NewTimer(elevator.config.doorOpenDuration_s)
		}
		elevator.requests[btn_floor][btn_type] = 1
	case EB_Moving:
		elevator.requests[btn_floor][btn_type] = 1
		break
	case EB_Idle:
		elevator.requests[btn_floor][btn_type] = 1
		a := requests_nextAction(elevator)
		elevator.dirn = a.dirn
		elevator.behaviour = a.behaviour
		switch elevator.behaviour {
		case EB_DoorOpen:
			io_setDoorOpenLamp(true)
			door_timer = time.NewTimer(elevator.config.doorOpenDuration_s)
			elevator = requests_clearAtCurrentFloor(elevator)
		case EB_Moving:
			io_setMotorDirection(elevator.dirn)
		case EB_Idle:

		}

	}
	setAllLights(elevator)

	println("\nNew state:\n")
	elevator_print(elevator)
}

func fsm_onFloorArrival(newFloor int) {
	//fmt.Println("\n\n%s(%d)\n", __FUNCTION__, newFloor)
	elevator_print(elevator)

	elevator.floor = newFloor

	io_setFloorIndicator(newFloor)

	switch elevator.behaviour {
	case EB_Moving:
		if requests_shouldStop(elevator) {
			io_setMotorDirection(D_Stop)
			io_setDoorOpenLamp(true)
			elevator := requests_clearAtCurrentFloor(elevator)
			door_timer = time.NewTimer(elevator.config.doorOpenDuration_s)
			setAllLights(elevator)
			elevator.behaviour = EB_DoorOpen
		}
	default:
	}

	fmt.Println("New state:")
	elevator_print(elevator)
}

func fsm_onDoorTimeout() {
	//fmt.Println("\n\n%s()\n", __FUNCTION__)
	//elevator_print(elevator)

	switch elevator.behaviour {
	case EB_DoorOpen:
		a := requests_nextAction(elevator)
		elevator.dirn = a.dirn
		elevator.behaviour = a.behaviour

		switch elevator.behaviour {
		case EB_DoorOpen:
			door_timer = time.NewTimer(elevator.config.doorOpenDuration_s)
			elevator := requests_clearAtCurrentFloor(elevator)
			setAllLights(elevator)
		case EB_Moving:
			io_setDoorOpenLamp(false)
			io_setMotorDirection(elevator.dirn)
		case EB_Idle:
			io_setDoorOpenLamp(false)
			io_setMotorDirection(elevator.dirn)
		}
	default:
	}
	//fmt.Println("\nNew state:\n")
	//elevator_print(elevator)
}
