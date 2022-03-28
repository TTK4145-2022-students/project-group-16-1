package elevator_control

import (
	. "Elevator-project/src/constants"
	"Elevator-project/src/elevio"
	"time"
)

func ElevatorControl(
	oa_ec_assignedOrders <-chan [N_FLOORS][N_BTN_TYPES]bool,
	drv_ec_floor <-chan int,
	drv_ec_obstr <-chan bool,
	drv_ec_stop <-chan bool,
	ec_net_elevatorState chan<- ElevatorState,
	ec_oa_localElevatorState chan<- ElevatorState,
	ec_or_localOrderServed chan<- elevio.ButtonEvent,
	id string) {

	elevator := &Elevator{}
	elevator_init(elevator)

	door_timer := time.NewTimer(time.Second)
	door_timer.Stop()
	too_late_timer := time.NewTimer(IS_FUCKED_DURATION)
	too_late_timer.Stop()
	too_late_timer_ticking := false
	send_state_timeout := time.NewTicker(INTERVAL)

	for {
		select {
		case assigned_orders := <-oa_ec_assignedOrders:
			elevator.requests = assigned_orders
			update_too_late_timer(elevator, too_late_timer, &too_late_timer_ticking)

			switch elevator.behaviour {
			case EB_DoorOpen:
				should_clear_btns := requests_shouldClearImmediately(elevator)
				for _, val := range should_clear_btns {
					ec_or_localOrderServed <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.ButtonType(val)}

				}

			case EB_Idle:
				next_action := requests_nextAction(elevator)

				elevator.dirn = next_action.dirn
				elevator.behaviour = next_action.behaviour
				switch elevator.behaviour {
				case EB_DoorOpen:
					io_setDoorOpenLamp(true)
					door_timer.Reset(DOOR_OPEN_DURATION)
					should_clear_btns := requests_clearAtCurrentFloor(elevator)
					for _, val := range should_clear_btns {
						ec_or_localOrderServed <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.ButtonType(val)}

					}
				case EB_Moving:
					io_setMotorDirection(elevator.dirn)
				case EB_Idle:

				}

			}

		case new_floor := <-drv_ec_floor:

			too_late_timer.Stop()
			too_late_timer_ticking = false
			elevator.too_late = false

			elevator.floor = new_floor

			io_setFloorIndicator(new_floor)

			switch elevator.behaviour {
			case EB_Moving:

				if requests_shouldStop(elevator) {

					io_setMotorDirection(D_Stop)
					io_setDoorOpenLamp(true)
					should_clear_btns := requests_clearAtCurrentFloor(elevator)
					for _, val := range should_clear_btns {
						ec_or_localOrderServed <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.ButtonType(val)}

					}
					door_timer.Reset(DOOR_OPEN_DURATION)
					elevator.behaviour = EB_DoorOpen
				}
			default:
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
					io_setDoorOpenLamp(false)
					io_setMotorDirection(elevator.dirn)
				}
			}

		case <-send_state_timeout.C:
			elevator_state := createElevatorStateMSG(elevator, id)
			ec_net_elevatorState <- elevator_state
			ec_oa_localElevatorState <- elevator_state

		case <-too_late_timer.C:
			elevator.too_late = true
		}
	}
}

func update_too_late_timer(
	elevator *Elevator,
	too_late_timer *time.Timer,
	too_late_timer_ticking *bool) {

	no_local_orders := true
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BTN_TYPES; btn++ {
			if elevator.requests[floor][btn] {
				no_local_orders = false
				if !(floor == elevator.floor) {
					if !*too_late_timer_ticking {
						too_late_timer.Reset(IS_FUCKED_DURATION)
						*too_late_timer_ticking = true
					}
				}
			}
		}
	}
	if no_local_orders {
		too_late_timer.Stop()
		*too_late_timer_ticking = false
		elevator.too_late = false
	}
}
