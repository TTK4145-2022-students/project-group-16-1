package elevator_control

import (
	"Elevator-project/src/elevio"
	"time"
)

func Elevator_control(drv_buttons chan elevio.ButtonEvent, drv_floors chan int, drv_obstr chan bool, drv_stop chan bool) {
	println("Elevator control started!")

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

			for f := 0; f < N_FLOORS; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		case <-door_timer.C:
			fsm_onDoorTimeout()
		}
	}
}
