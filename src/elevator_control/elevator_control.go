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
			fsm_onRequestButtonPress(a)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			fsm_onFloorArrival(a.Floor)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}
}
