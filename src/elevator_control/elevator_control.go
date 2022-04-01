package elevator_control

import (
	. "Elevator-project/src/constants"
	"Elevator-project/src/elevator_io"
	"fmt"
	"time"
)

func ElevatorControl(
	oa_ec_assignedOrders 	<-chan [N_FLOORS][N_BTN_TYPES]bool,
	eio_ec_floor 			<-chan int,
	eio_ec_obstr 			<-chan bool,
	ec_net_elevatorStateTx 	chan<- ElevatorStateMsg,
	ec_oa_elevatorState 	chan<- ElevatorStateMsg,
	ec_or_localOrderServed 	chan<- elevator_io.ButtonEvent,
	id 						string,
	) {
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
			if elevator.motor_failure {
				break
			}
			elevator.orders = assigned_orders
			handle_motor_failure_timer(elevator, motor_failure_timer, &motor_failure_timer_ticking)

			switch elevator.behaviour {
			case EB_DoorOpen:
				btns_to_clear := orders_shouldClearImmediately(elevator)
				for _, btn := range btns_to_clear {
					ec_or_localOrderServed <- elevator_io.ButtonEvent{Floor: elevator.floor, Button: elevator_io.ButtonType(btn)}
				}

			case EB_Idle:
				next_action := orders_nextAction(elevator)
				elevator.dirn = next_action.dirn
				elevator.behaviour = next_action.behaviour

				switch elevator.behaviour {
				case EB_DoorOpen:
					elevator_io.SetDoorOpenLamp(true)
					door_timer.Reset(DOOR_OPEN_DURATION)
					btns_to_clear := orders_clearAtCurrentFloor(elevator)
					for _, btn := range btns_to_clear {
						ec_or_localOrderServed <- elevator_io.ButtonEvent{Floor: elevator.floor, Button: elevator_io.ButtonType(btn)}
					}
				case EB_Moving:
					elevator_io.SetMotorDirection(elevator.dirn)
				case EB_Idle:
				}
			}

		case new_floor := <-eio_ec_floor:
			//Arrived at new florr => motor OK
			motor_failure_timer.Stop()
			motor_failure_timer_ticking = false
			elevator.motor_failure = false

			elevator.floor = new_floor
			elevator_io.SetFloorIndicator(new_floor)

			switch elevator.behaviour {
			case EB_Moving:
				if orders_shouldStop(elevator) {
					elevator_io.SetMotorDirection(elevator_io.MD_Stop)
					elevator_io.SetDoorOpenLamp(true)
					btns_to_clear := orders_clearAtCurrentFloor(elevator)
					for _, btn := range btns_to_clear {
						ec_or_localOrderServed <- elevator_io.ButtonEvent{Floor: elevator.floor, Button: elevator_io.ButtonType(btn)}
					}
					door_timer.Reset(DOOR_OPEN_DURATION)
					elevator.behaviour = EB_DoorOpen
				}
			}

		case obstructed := <-eio_ec_obstr:
			elevator.obstructed = obstructed

		case <-door_timer.C:

			switch elevator.behaviour {
			case EB_DoorOpen:
				if elevator.obstructed {
					door_timer.Reset(DOOR_OPEN_DURATION)
					break
				}

				next_action := orders_nextAction(elevator)
				elevator.dirn = next_action.dirn
				elevator.behaviour = next_action.behaviour

				switch elevator.behaviour {
				case EB_DoorOpen:
					door_timer.Reset(DOOR_OPEN_DURATION)
					orders_clearAtCurrentFloor(elevator)
				case EB_Idle, EB_Moving:
					elevator_io.SetDoorOpenLamp(false)
					elevator_io.SetMotorDirection(elevator.dirn)
				}
			}

		case <-send_state_ticker.C:
			elevator_msg := createElevatorStateMsg(elevator, id)
			ec_net_elevatorStateTx <- elevator_msg
			ec_oa_elevatorState <- elevator_msg

		case <-motor_failure_timer.C:
			elevator.motor_failure = true
			motor_failure_timer_ticking = false
			fmt.Println("Timeout - motor failure")
		}
	}
}

func elevator_init(elevator *Elevator) {
	fmt.Println("Initialising FSM: ")
	*elevator = createUninitialisedElevator()
	if elevator.floor == -1 {
		elevator_io.SetMotorDirection(elevator_io.MD_Down)
		elevator.dirn = elevator_io.MD_Down
		elevator.behaviour = EB_Moving
	}

}

// Start motor failure timer if local assigned orders on other floor.
// If no local asigned orders, motor has not failed (for our purposes)
func handle_motor_failure_timer(
	elevator 					*Elevator,
	motor_failure_timer 		*time.Timer,
	motor_failure_timer_ticking *bool,
	) {

	no_local_orders := true
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BTN_TYPES; btn++ {
			if elevator.orders[floor][btn] {
				no_local_orders = false	
				if !*motor_failure_timer_ticking {
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
