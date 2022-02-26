package elevator_control

import (
	"time"
)

var elevator *Elevator
var door_timer *time.Timer

const N_FLOORS = 4 //REMOVE THIS
const N_BUTTONS = 3
const HARDWARE_ADDR = "localhost:15657"

func ElevatorControl(assigner_assignedOrders chan [N_FLOORS][N_BUTTONS]bool, drv_floors chan int, drv_obstr chan bool, drv_stop chan bool) {
	println("Elevator control started!")

	// Needed for initing timer!
	door_timer = time.NewTimer(time.Second)
	door_timer.Stop()

	fsm_init()

	for {
		select {
		case a := <-assigner_assignedOrders:

			fsm_onRequestUpdate(a)

		case a := <-drv_floors:

			fsm_onFloorArrival(a)

		case a := <-drv_obstr:
			fsm_onObstruction(a)

		case <-drv_stop:

		case <-door_timer.C:
			fsm_onDoorTimeout()
		}
	}
}
