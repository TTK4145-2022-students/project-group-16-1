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
	

	door_timer := time.NewTimer(time.Second)
	door_timer.Stop()
	motor_failure_timer := time.NewTimer(MOTOR_FAILURE_DURATION)
	send_state_ticker := time.NewTicker(PERIODIC_SEND_DURATION)

	elevator := &Elevator{}
	elevator_init(elevator)

	for {
		select {
		case assigned_orders := <-oa_ec_assignedOrders:
			if elevator.motor_failure {
				break
			}

			elevator.orders = assigned_orders
			
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
					motor_failure_timer.Reset(MOTOR_FAILURE_DURATION)
				case EB_Idle:
				}
			}

		case new_floor := <-eio_ec_floor:
			//Arrived at new florr => motor OK
			motor_failure_timer.Stop()
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
				} else {
					motor_failure_timer.Reset(MOTOR_FAILURE_DURATION)
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
				case EB_Idle:
					elevator_io.SetDoorOpenLamp(false)
					elevator_io.SetMotorDirection(elevator.dirn)
				case EB_Moving:
					elevator_io.SetDoorOpenLamp(false)
					elevator_io.SetMotorDirection(elevator.dirn)
					motor_failure_timer.Reset(MOTOR_FAILURE_DURATION)
				}
			}

		case <-send_state_ticker.C:
			elevator_msg := createElevatorStateMsg(elevator, id)
			ec_net_elevatorStateTx <- elevator_msg
			ec_oa_elevatorState <- elevator_msg

		case <-motor_failure_timer.C:
			elevator.motor_failure = true
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
