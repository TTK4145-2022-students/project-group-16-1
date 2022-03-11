package elevator_control

import (
	"Elevator-project/src/elevio"
	"time"
)

var elevator *Elevator
var door_timer *time.Timer
var id string

const N_FLOORS = 4 //REMOVE THIS
const N_BUTTONS = 3
const HARDWARE_ADDR = "localhost:15657"
const INTERVAL = 50 * time.Millisecond

func ElevatorControl(
	oa_assignedOrders <-chan [N_FLOORS][N_BUTTONS]bool,
	drv_floors <-chan int,
	drv_obstr <-chan bool,
	drv_stop <-chan bool,
	net_elevatorState chan<- ElevatorState,
	ec_localOrderServed chan<- elevio.ButtonEvent) {
	println("Elevator control started!")

	// Needed for initing timer!
	door_timer = time.NewTimer(time.Second)
	door_timer.Stop()

	fsm_init()
	send_state_timeout := time.NewTimer(INTERVAL)

	for {
		select {
		case a := <-oa_assignedOrders:

			fsm_onRequestUpdate(a)

		case a := <-drv_floors:

			fsm_onFloorArrival(a)

		case a := <-drv_obstr:
			fsm_onObstruction(a)

		case <-drv_stop:

		case <-door_timer.C:
			fsm_onDoorTimeout()
		case <-send_state_timeout.C:
			elevator_state := createElevatorStateMSG()
			net_elevatorState <- elevator_state
		}
	}
}
