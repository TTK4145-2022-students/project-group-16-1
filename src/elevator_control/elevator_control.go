package elevator_control

import (
	"Elevator-project/src/elevio"
	"fmt"
	"time"
)

var elevator *Elevator
var door_timer *time.Timer
var id string

const N_FLOORS = 4 //REMOVE THIS
const N_BTN_TYPES = 3
const HARDWARE_ADDR = "localhost:15657"
const INTERVAL = 50 * time.Millisecond

func ElevatorControl(
	oa_ec_assignedOrders <-chan [N_FLOORS][N_BTN_TYPES]bool,
	drv_ec_floor <-chan int,
	drv_ec_obstr <-chan bool,
	drv_ec_stop <-chan bool,
	net_elevatorState chan<- ElevatorState,
	ec_or_localOrderServed chan<- elevio.ButtonEvent,
	local_id string) {
	println("Elevator control started!")
	id = local_id
	// Needed for initing timer!
	door_timer = time.NewTimer(time.Second)
	door_timer.Stop()

	fsm_init()
	send_state_timeout := time.After(INTERVAL)

	for {
		select {
		case a := <-oa_ec_assignedOrders:
			fmt.Println("--------------")
			fmt.Println("Jumping into [fsm_onRequestUpdate]")
			elevator_print(elevator)

			elevator.requests = a
			switch elevator.behaviour {
			case EB_DoorOpen, EB_Obstructed:
				should_clear_btns := requests_shouldClearImmediately(elevator)
				for _, val := range should_clear_btns {
					ec_or_localOrderServed <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.ButtonType(val)}

				}

			case EB_Idle:
				a := requests_nextAction(elevator)
				elevator.dirn = a.dirn
				elevator.behaviour = a.behaviour
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
			setAllLights(elevator)

			fmt.Println("New state:")
			elevator_print(elevator)
			fmt.Println("--------------")

		case new_floor := <-drv_ec_floor:

			fmt.Println("--------------")
			fmt.Println("Jumping into [fsm_onFloorArrival]")
			elevator_print(elevator)

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
					setAllLights(elevator)
					elevator.behaviour = EB_DoorOpen
				}
			default:
			}

			fmt.Println("New state:")
			elevator_print(elevator)
			fmt.Println("--------------")

		case a := <-drv_ec_obstr:
			fsm_onObstruction(a)

		case <-drv_ec_stop:

		case <-door_timer.C:
			fsm_onDoorTimeout()
		case <-send_state_timeout:
			elevator_state := createElevatorStateMSG()
			net_elevatorState <- elevator_state
			send_state_timeout = time.After(INTERVAL)
		}
	}
}
