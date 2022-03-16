package elevator_control

import (
	"Elevator-project/src/elevio"
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

			fsm_onRequestUpdate(a)

		case a := <-drv_ec_floor:

			fsm_onFloorArrival(a)

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
