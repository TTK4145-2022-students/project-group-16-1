package elevator_control

import (
	"Elevator-project/src/elevio"
	"time"
)

var elevator *Elevator
var door_timer *time.Timer

const N_FLOORS = 4 //REMOVE THIS
const N_BUTTONS = 3
const HARDWARE_ADDR = "localhost:15657"

func Elevator_control(drv_buttons chan elevio.ButtonEvent, drv_floors chan int, drv_obstr chan bool, drv_stop chan bool) {
	println("Elevator control started!")

	// Needed for initing timer!
	door_timer = time.NewTimer(time.Second)
	door_timer.Stop()

	fsm_init()

	for {
		select {
		case a := <-drv_buttons:

			btn_floor := a.Floor
			btn_type := Button(a.Button)
			fsm_onRequestButtonPress(btn_floor, btn_type)

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
