package elevator_control

import (
	"Elevator-project/src/elevio"
	"fmt"
)

func Elevator_control(numFloors int, inputPollRate_ms float64, drv_buttons chan elevio.ButtonEvent, drv_floors chan int, drv_obstr chan bool, drv_stop chan bool) {
	println("Elevator control started!")

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			btn_floor := a.Floor
			btn_type := Button(a.Button)
			fsm_onRequestButtonPress(btn_floor, btn_type)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			fsm_onFloorArrival(a)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				io_setMotorDirection(D_Stop)
			} else {
				io_setMotorDirection(elevator.dirn)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		case <-door_timer.C:
			fsm_onDoorTimeout()
		}
	}
}
