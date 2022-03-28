package elevator_control

import (
	. "Elevator-project/src/constants"
	"Elevator-project/src/elevio"
	"fmt"
	"time"
)

func ElevatorControl(
	oa_ec_assignedOrders <-chan [N_FLOORS][N_BTN_TYPES]bool,
	drv_ec_floor <-chan int,
	drv_ec_obstr <-chan bool,
	drv_ec_stop <-chan bool,
	ec_net_elevator chan<- ElevatorMsg,
	ec_oa_elevator chan<- ElevatorMsg,
	ec_or_localOrderServed chan<- elevio.ButtonEvent,
	id string) {

	elevator := &Elevator{}
	elevator_init(elevator)

	door_timer := time.NewTimer(time.Second)
	door_timer.Stop()
	motor_failure_timer := time.NewTimer(MOTOR_FAILURE_DURATION)
	motor_failure_timer.Stop()
	motor_failure_timer_ticking := false
	send_state_ticker := time.NewTicker(PERIODIC_SEND_DURATION)

	for {
		select {
		case assigned_orders := <-oa_ec_assignedOrders:
			elevator.requests = assigned_orders
			handle_motor_failure_timer(elevator, motor_failure_timer, &motor_failure_timer_ticking)

			switch elevator.behaviour {
			case EB_DoorOpen:
				btns_to_clear := requests_shouldClearImmediately(elevator)
				for _, btn := range btns_to_clear {
					ec_or_localOrderServed <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.ButtonType(btn)}
				}

			case EB_Idle:
				next_action := requests_nextAction(elevator)
				elevator.dirn = next_action.dirn
				elevator.behaviour = next_action.behaviour

				switch elevator.behaviour {
				case EB_DoorOpen:
					elevio.SetDoorOpenLamp(true)
					door_timer.Reset(DOOR_OPEN_DURATION)
					btns_to_clear := requests_clearAtCurrentFloor(elevator)
					for _, btn := range btns_to_clear {
						ec_or_localOrderServed <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.ButtonType(btn)}
					}
				case EB_Moving:
					elevio.SetMotorDirection(elevator.dirn)
				case EB_Idle:
				}
			}

		case new_floor := <-drv_ec_floor:
			//motor OK
			motor_failure_timer.Stop()
			motor_failure_timer_ticking = false
			elevator.motor_failure = false

			elevator.floor = new_floor
			elevio.SetFloorIndicator(new_floor)

			switch elevator.behaviour {
			case EB_Moving:
				if requests_shouldStop(elevator) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevio.SetDoorOpenLamp(true)
					btns_to_clear := requests_clearAtCurrentFloor(elevator)
					for _, btn := range btns_to_clear {
						ec_or_localOrderServed <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.ButtonType(btn)}
					}
					door_timer.Reset(DOOR_OPEN_DURATION)
					elevator.behaviour = EB_DoorOpen
				}
			}

		case obstructed := <-drv_ec_obstr:
			elevator.obstructed = obstructed

		case <-door_timer.C:

			switch elevator.behaviour {
			case EB_DoorOpen:
				if elevator.obstructed {
					door_timer.Reset(DOOR_OPEN_DURATION)
					break
				}

				next_action := requests_nextAction(elevator)
				elevator.dirn = next_action.dirn
				elevator.behaviour = next_action.behaviour

				switch elevator.behaviour {
				case EB_DoorOpen:
					door_timer.Reset(DOOR_OPEN_DURATION)
					requests_clearAtCurrentFloor(elevator)
				case EB_Idle, EB_Moving:
					elevio.SetDoorOpenLamp(false)
					elevio.SetMotorDirection(elevator.dirn)
				}
			}

		case <-send_state_ticker.C:
			elevator_msg := createElevatorMsg(elevator, id)
			ec_net_elevator <- elevator_msg
			ec_oa_elevator <- elevator_msg

		case <-motor_failure_timer.C:
			elevator.motor_failure = true
		}
	}
}

func elevator_init(elevator *Elevator) {
	fmt.Println("Initialising FSM: ")
	*elevator = elevator_uninitialised()
	if elevator.floor == -1 {
		elevio.SetMotorDirection(elevio.MD_Down)
		elevator.dirn = elevio.MD_Down
		elevator.behaviour = EB_Moving
	}

}

// Start motor failure timer if local assigned orders on other floor.
// If no local asigned orders, motor has not failed (for our purposes)
func handle_motor_failure_timer(
	elevator *Elevator,
	motor_failure_timer *time.Timer,
	motor_failure_timer_ticking *bool) {

	no_local_orders := true
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BTN_TYPES; btn++ {
			if elevator.requests[floor][btn] {
				no_local_orders = false
				if !(floor == elevator.floor) && !*motor_failure_timer_ticking {
					motor_failure_timer.Reset(MOTOR_FAILURE_DURATION)
					*motor_failure_timer_ticking = true
				}
			}
		}
	}

	if no_local_orders {
		motor_failure_timer.Stop()
		*motor_failure_timer_ticking = false
		elevator.motor_failure = false
	}
}
