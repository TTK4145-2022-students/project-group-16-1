package elevator_control

import (
	"Elevator-project/src/elevio"
	"fmt"
	"time"
)

const N_FLOORS = 4 //REMOVE THIS
const N_BTN_TYPES = 3
const HARDWARE_ADDR = "localhost:15657"
const INTERVAL = 50 * time.Millisecond

func ElevatorControl(
	oa_ec_assignedOrders <-chan [N_FLOORS][N_BTN_TYPES]bool,
	drv_ec_floor <-chan int,
	drv_ec_obstr <-chan bool,
	drv_ec_stop <-chan bool,
	net_ec_elevatorState chan<- ElevatorState,
	ec_oa_localElevatorState chan<- ElevatorState,
	ec_or_localOrderServed chan<- elevio.ButtonEvent,
	local_id string) {
	println("Elevator control started!")

	elevator := &Elevator{}
	id := local_id
	var door_timer *time.Timer
	var is_fucked_timer *time.Timer
	hardware_timer_stopped := true
	// Needed for initing timer!
	door_timer = time.NewTimer(time.Second)
	door_timer.Stop()
	is_fucked_timer = time.NewTimer(time.Second * 15)
	if !is_fucked_timer.Stop() {
		<-is_fucked_timer.C
	}

	elevator_init(elevator)
	elevator_print(elevator)
	send_state_timeout := time.After(INTERVAL)

	for {
		fmt.Println(elevator.too_late)
		select {
		case assigned_orders := <-oa_ec_assignedOrders:
			fmt.Println("--------------")
			fmt.Println("Jumping into [onRequestUpdate]")
			elevator_print(elevator)

			elevator.requests = assigned_orders

			if hardware_timer_stopped {
				for floor := 0; floor < N_FLOORS; floor++ {
					for btn := 0; btn < N_BTN_TYPES; btn++ {
						if assigned_orders[floor][btn] {
							if !(floor == elevator.floor) {
								is_fucked_timer.Reset(time.Second * 15)
								hardware_timer_stopped = false
							}
						}
					}
				}
			}
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
					door_timer.Reset(elevator.config.doorOpenDuration_s)
					should_clear_btns := requests_clearAtCurrentFloor(elevator)
					for _, val := range should_clear_btns {
						ec_or_localOrderServed <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.ButtonType(val)}

					}
				case EB_Moving:
					io_setMotorDirection(elevator.dirn)
				case EB_Idle:

				}

			}

			fmt.Println("New state:")
			elevator_print(elevator)
			fmt.Println("--------------")

		case new_floor := <-drv_ec_floor:

			fmt.Println("--------------")
			fmt.Println("Jumping into [fsm_onFloorArrival]")
			elevator_print(elevator)

			is_fucked_timer.Stop()
			hardware_timer_stopped = true
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
					door_timer.Reset(elevator.config.doorOpenDuration_s)
					elevator.behaviour = EB_DoorOpen
				}
			default:
			}

			fmt.Println("New state:")
			elevator_print(elevator)
			fmt.Println("--------------")

		case obstructed := <-drv_ec_obstr:
			fmt.Println("--------------")
			fmt.Println("Jumping into [fsm_onObstruction")
			elevator_print(elevator)
			elevator.obstructed = obstructed
			fmt.Println("New state:")
			elevator_print(elevator)
			fmt.Println("--------------")

		case <-door_timer.C:
			fmt.Println("Jumping into [fsm_onDoorTimeout]")
			elevator_print(elevator)
			switch elevator.behaviour {
			case EB_DoorOpen:
				if elevator.obstructed {
					door_timer.Reset(elevator.config.doorOpenDuration_s)
					break
				}
				next_action := requests_nextAction(elevator)
				elevator.dirn = next_action.dirn
				elevator.behaviour = next_action.behaviour

				switch elevator.behaviour {
				case EB_DoorOpen:
					door_timer.Reset(elevator.config.doorOpenDuration_s)
					requests_clearAtCurrentFloor(elevator)
				case EB_Idle, EB_Moving:
					io_setDoorOpenLamp(false)
					io_setMotorDirection(elevator.dirn)
				}
			}
			fmt.Println("New state:")
			elevator_print(elevator)
			fmt.Println("--------------")

		case <-send_state_timeout:
			elevator_state := createElevatorStateMSG(elevator, id)
			net_ec_elevatorState <- elevator_state
			ec_oa_localElevatorState <- elevator_state
			send_state_timeout = time.After(INTERVAL)

		case <-is_fucked_timer.C:
			elevator.too_late = true
		}
	}
}
