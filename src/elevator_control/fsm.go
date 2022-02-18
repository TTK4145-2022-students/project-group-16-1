package elevator_control

import (
	"Elevator-project/src/elevio"
)

var elevator Elevator

const N_FLOORS = 4 //REMOVE THIS
const N_BUTTONS = 3

// func fsm_init(){ //returnerer void __attribute__((constructor))
// 	elevator := elevator_uninitialized()

//     con_load("elevator.con",
//         con_val("doorOpenDuration_s", &elevator.config.doorOpenDuration_s, "%lf")
//         con_enum("clearRequestVariant", &elevator.config.clearRequestVariant,
//             con_match(CV_All)
//             con_match(CV_InDirn)
//         )
//     )
// 	outputDevice := elevio_getOutputDevice()
// }

func setAllLights(es Elevator) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			elevio.SetButtonLamp(btn, floor, elevator.requests[floor][btn] != 0)
		}
	}
}

func fsm_onInitBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.dirn = D_Down
	elevator.behaviour = EB_Moving
}

func fsm_onRequestButtonPress(btn_floor int, btn_type Button) {
	Println("\n\n%s(%d, %s)\n", __FUNCTION__, btn_floor, elevio_button_toString(btn_type))
	elevator_print(elevator)
	switch elevator.behaviour {
	case EB_DoorOpen:
		//IF requests_shouldClearImmediately:
		//		timer_start()
		//ELSE:
		elevator.requests[btn_floor][btn_type] = 1
		break
	case EB_Moving:
		elevator.requests[btn_floor][btn_type] = 1
		break
	case EB_Idle:
		elevator.requests[btn_floor][btn_type] = 1
		//calculate next action
		//update behaviour and direction accordingly
		elevator.behaviour = elevator.behaviour
		switch elevator.behaviour { //switch on new updated elevator behaviour
		case EB_DoorOpen:
			//door light
			//time start
			//clear requests at current floor (elevator = request_ClearAtCurrentFloor())
			break
		case EB_Moving:
			//Motor direction = elevator.dirn
			break
		case EB_Idle:

			break
		}
		break
	}
	setAllLights(elevator)
	println("\nNew state:\n")
	elevator_print(elevator)
}

func fsm_onFloorArrival(newFloor int) {
	//fmt.Println("\n\n%s(%d)\n", __FUNCTION__, newFloor)
	//elevator_print(elevator)

	elevator.floor = newFloor

	elevio.SetFloorIndicator(newFloor)

	switch elevator.behaviour {
	case EB_Moving:
		if requests_shouldStop(elevator) {
			elevio.SetMotorDirection(D_Stop)
			elevio.SetDoorOpenLamp(true)
			elevator := requests_clearAtCurrentFloor(elevator)
			timer_start(elevator.config.doorOpenDuration_s)
			setAllLights(elevator)
			elevator.behaviour = EB_DoorOpen
		}
		break
	default:
		break
	}

	//fmt.Println("\nNew state:\n")
	//elevator_print(elevator)
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
			//timer_start(elevator.config.doorOpenDuration_s)
			elevator := requests_clearAtCurrentFloor(elevator)
			setAllLights(elevator)
			break
		case EB_Moving:
		case EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevator.dirn)
			break
		}
		break
	default:
		break
	}
	//fmt.Println("\nNew state:\n")
	//elevator_print(elevator)
}
