package elevator_control

import (
	. "Elevator-project/src/constants"
	"Elevator-project/src/elevio"
)

type Action struct {
	dirn      elevio.MotorDirection
	behaviour ElevatorBehaviour
}

func orders_above(e *Elevator) bool {
	for f := e.floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BTN_TYPES; btn++ {
			if e.orders[f][btn] {
				return true
			}
		}
	}
	return false
}

func orders_below(e *Elevator) bool {
	for f := 0; f < e.floor; f++ {
		for btn := 0; btn < N_BTN_TYPES; btn++ {
			if e.orders[f][btn] {
				return true
			}
		}
	}
	return false
}

func orders_here(e *Elevator) bool {
	for btn := 0; btn < N_BTN_TYPES; btn++ {
		if e.orders[e.floor][btn] {
			return true
		}
	}
	return false
}

func orders_nextAction(e *Elevator) Action {
	switch e.dirn {
	case elevio.MD_Up:
		var action Action

		if orders_above(e) {
			action.behaviour = EB_Moving
			action.dirn = elevio.MD_Up
		} else if orders_here(e) {
			action.behaviour = EB_DoorOpen
			action.dirn = elevio.MD_Down
		} else if orders_below(e) {
			action.behaviour = EB_Moving
			action.dirn = elevio.MD_Down
		} else {
			action.behaviour = EB_Idle
			action.dirn = elevio.MD_Stop
		}

		return action

	case elevio.MD_Down:
		var action Action

		if orders_below(e) {
			action.behaviour = EB_Moving
			action.dirn = elevio.MD_Down
		} else if orders_here(e) {
			action.behaviour = EB_DoorOpen
			action.dirn = elevio.MD_Up
		} else if orders_above(e) {
			action.behaviour = EB_Moving
			action.dirn = elevio.MD_Up
		} else {
			action.behaviour = EB_Idle
			action.dirn = elevio.MD_Stop
		}

		return action
	case elevio.MD_Stop: // there should only be one order in the Stop case. Checking up or down first is arbitrary.
		var action Action

		if orders_here(e) {
			action.behaviour = EB_DoorOpen
			action.dirn = elevio.MD_Stop
		} else if orders_below(e) {
			action.behaviour = EB_Moving
			action.dirn = elevio.MD_Down
		} else if orders_above(e) {
			action.behaviour = EB_Moving
			action.dirn = elevio.MD_Up
		} else {

			action.behaviour = EB_Idle
			action.dirn = elevio.MD_Stop
		}
		return action

	default:
		var action Action
		action.behaviour = EB_Idle
		action.dirn = elevio.MD_Stop
		return action
	}
}

func orders_shouldStop(e *Elevator) bool {
	switch e.dirn {
	case elevio.MD_Down:
		return (e.orders[e.floor][elevio.BT_HallDown]) || (e.orders[e.floor][elevio.BT_Cab]) || !orders_below(e)
	case elevio.MD_Up:
		return (e.orders[e.floor][elevio.BT_HallUp]) || (e.orders[e.floor][elevio.BT_Cab]) || !orders_above(e)
	case elevio.MD_Stop:
		return true
	default:
		return true
	}
}

func orders_shouldClearImmediately(e *Elevator) []elevio.ButtonType {
	btns_to_clear := []elevio.ButtonType{}

	if e.dirn == elevio.MD_Up {
		if e.orders[e.floor][elevio.BT_HallUp] == true {
			e.orders[e.floor][elevio.BT_HallUp] = false
			btns_to_clear = append(btns_to_clear, elevio.BT_HallUp)
		}
	} else if e.dirn == elevio.MD_Down {
		if e.orders[e.floor][elevio.BT_HallDown] == true {
			e.orders[e.floor][elevio.BT_HallDown] = false
			btns_to_clear = append(btns_to_clear, elevio.BT_HallDown)
		}
	} else {
		if e.orders[e.floor][elevio.BT_HallUp] == true {
			e.orders[e.floor][elevio.BT_HallUp] = false
			btns_to_clear = append(btns_to_clear, elevio.BT_HallUp)
			e.dirn = elevio.MD_Up
		} else if e.orders[e.floor][elevio.BT_HallDown] == true {
			e.orders[e.floor][elevio.BT_HallDown] = false
			btns_to_clear = append(btns_to_clear, elevio.BT_HallDown)
			e.dirn = elevio.MD_Down
		}
	}
	if e.orders[e.floor][elevio.BT_Cab] == true {
		e.orders[e.floor][elevio.BT_Cab] = false
		btns_to_clear = append(btns_to_clear, elevio.BT_Cab)
	}

	return btns_to_clear
}

func orders_clearAtCurrentFloor(e *Elevator) []elevio.ButtonType {
	btns_to_clear := []elevio.ButtonType{}
	if e.orders[e.floor][elevio.BT_Cab] == true {
		e.orders[e.floor][elevio.BT_Cab] = false
		btns_to_clear = append(btns_to_clear, elevio.BT_Cab)
	}

	switch e.dirn {
	case elevio.MD_Up:
		if !orders_above(e) && (!e.orders[e.floor][elevio.BT_HallUp]) {
			if e.orders[e.floor][elevio.BT_HallDown] == true {
				e.orders[e.floor][elevio.BT_HallDown] = false
				btns_to_clear = append(btns_to_clear, elevio.BT_HallDown)
			}
		}
		if e.orders[e.floor][elevio.BT_HallUp] == true {
			e.orders[e.floor][elevio.BT_HallUp] = false
			btns_to_clear = append(btns_to_clear, elevio.BT_HallUp)
		}

	case elevio.MD_Down:
		if !orders_below(e) && (!e.orders[e.floor][elevio.BT_HallDown]) {

			if e.orders[e.floor][elevio.BT_HallUp] == true {
				e.orders[e.floor][elevio.BT_HallUp] = false
				btns_to_clear = append(btns_to_clear, elevio.BT_HallUp)
			}
		}
		if e.orders[e.floor][elevio.BT_HallDown] == true {
			e.orders[e.floor][elevio.BT_HallDown] = false
			btns_to_clear = append(btns_to_clear, elevio.BT_HallDown)
		}

	case elevio.MD_Stop:
		if e.orders[e.floor][elevio.BT_HallUp] == true {
			e.orders[e.floor][elevio.BT_HallUp] = false
			btns_to_clear = append(btns_to_clear, elevio.BT_HallUp)
		} else if e.orders[e.floor][elevio.BT_HallDown] == true {
			e.orders[e.floor][elevio.BT_HallDown] = false
			btns_to_clear = append(btns_to_clear, elevio.BT_HallDown)
		}
	}

	return btns_to_clear
}
