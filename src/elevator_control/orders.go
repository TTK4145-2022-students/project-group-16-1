package elevator_control

import (
	. "Elevator-project/src/constants"
	"Elevator-project/src/elevator_io"
)

type Action struct {
	dirn      elevator_io.MotorDirection
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
	case elevator_io.MD_Up:
		var action Action

		if orders_above(e) {
			action.behaviour = EB_Moving
			action.dirn = elevator_io.MD_Up
		} else if orders_here(e) {
			action.behaviour = EB_DoorOpen
			action.dirn = elevator_io.MD_Down
		} else if orders_below(e) {
			action.behaviour = EB_Moving
			action.dirn = elevator_io.MD_Down
		} else {
			action.behaviour = EB_Idle
			action.dirn = elevator_io.MD_Stop
		}

		return action

	case elevator_io.MD_Down:
		var action Action

		if orders_below(e) {
			action.behaviour = EB_Moving
			action.dirn = elevator_io.MD_Down
		} else if orders_here(e) {
			action.behaviour = EB_DoorOpen
			action.dirn = elevator_io.MD_Up
		} else if orders_above(e) {
			action.behaviour = EB_Moving
			action.dirn = elevator_io.MD_Up
		} else {
			action.behaviour = EB_Idle
			action.dirn = elevator_io.MD_Stop
		}

		return action
	case elevator_io.MD_Stop: // there should only be one order in the Stop case. Checking up or down first is arbitrary.
		var action Action

		if orders_here(e) {
			action.behaviour = EB_DoorOpen
			if e.orders[e.floor][elevator_io.BT_HallUp] {
				action.dirn = elevator_io.MD_Up
			} else if e.orders[e.floor][elevator_io.BT_HallDown] {
				action.dirn = elevator_io.MD_Down
			} else {
				action.dirn = elevator_io.MD_Stop
			}
		} else if orders_below(e) {
			action.behaviour = EB_Moving
			action.dirn = elevator_io.MD_Down
		} else if orders_above(e) {
			action.behaviour = EB_Moving
			action.dirn = elevator_io.MD_Up
		} else {

			action.behaviour = EB_Idle
			action.dirn = elevator_io.MD_Stop
		}
		return action

	default:
		var action Action
		action.behaviour = EB_Idle
		action.dirn = elevator_io.MD_Stop
		return action
	}
}

func orders_shouldStop(e *Elevator) bool {
	switch e.dirn {
	case elevator_io.MD_Down:
		return (e.orders[e.floor][elevator_io.BT_HallDown]) || (e.orders[e.floor][elevator_io.BT_Cab]) || !orders_below(e)
	case elevator_io.MD_Up:
		return (e.orders[e.floor][elevator_io.BT_HallUp]) || (e.orders[e.floor][elevator_io.BT_Cab]) || !orders_above(e)
	case elevator_io.MD_Stop:
		return true
	default:
		return true
	}
}

func orders_shouldClearImmediately(e *Elevator) []elevator_io.ButtonType {
	btns_to_clear := []elevator_io.ButtonType{}

	if e.dirn == elevator_io.MD_Up {
		if e.orders[e.floor][elevator_io.BT_HallUp]{
			e.orders[e.floor][elevator_io.BT_HallUp] = false
			btns_to_clear = append(btns_to_clear, elevator_io.BT_HallUp)
		}
	} else if e.dirn == elevator_io.MD_Down {
		if e.orders[e.floor][elevator_io.BT_HallDown]{
			e.orders[e.floor][elevator_io.BT_HallDown] = false
			btns_to_clear = append(btns_to_clear, elevator_io.BT_HallDown)
		}
	} else {
		if e.orders[e.floor][elevator_io.BT_HallUp]{
			e.orders[e.floor][elevator_io.BT_HallUp] = false
			btns_to_clear = append(btns_to_clear, elevator_io.BT_HallUp)
			e.dirn = elevator_io.MD_Up
		} else if e.orders[e.floor][elevator_io.BT_HallDown]{
			e.orders[e.floor][elevator_io.BT_HallDown] = false
			btns_to_clear = append(btns_to_clear, elevator_io.BT_HallDown)
			e.dirn = elevator_io.MD_Down
		}
	}
	if e.orders[e.floor][elevator_io.BT_Cab]{
		e.orders[e.floor][elevator_io.BT_Cab] = false
		btns_to_clear = append(btns_to_clear, elevator_io.BT_Cab)
	}

	return btns_to_clear
}

func orders_clearAtCurrentFloor(e *Elevator) []elevator_io.ButtonType {
	btns_to_clear := []elevator_io.ButtonType{}
	if e.orders[e.floor][elevator_io.BT_Cab]{
		e.orders[e.floor][elevator_io.BT_Cab] = false
		btns_to_clear = append(btns_to_clear, elevator_io.BT_Cab)
	}

	switch e.dirn {
	case elevator_io.MD_Up:
		if !orders_above(e) && (!e.orders[e.floor][elevator_io.BT_HallUp]) {
			if e.orders[e.floor][elevator_io.BT_HallDown]{
				e.orders[e.floor][elevator_io.BT_HallDown] = false
				btns_to_clear = append(btns_to_clear, elevator_io.BT_HallDown)
			}
		}
		if e.orders[e.floor][elevator_io.BT_HallUp]{
			e.orders[e.floor][elevator_io.BT_HallUp] = false
			btns_to_clear = append(btns_to_clear, elevator_io.BT_HallUp)
		}

	case elevator_io.MD_Down:
		if !orders_below(e) && (!e.orders[e.floor][elevator_io.BT_HallDown]) {

			if e.orders[e.floor][elevator_io.BT_HallUp]{
				e.orders[e.floor][elevator_io.BT_HallUp] = false
				btns_to_clear = append(btns_to_clear, elevator_io.BT_HallUp)
			}
		}
		if e.orders[e.floor][elevator_io.BT_HallDown]{
			e.orders[e.floor][elevator_io.BT_HallDown] = false
			btns_to_clear = append(btns_to_clear, elevator_io.BT_HallDown)
		}

	case elevator_io.MD_Stop:
		if e.orders[e.floor][elevator_io.BT_HallUp]{
			e.orders[e.floor][elevator_io.BT_HallUp] = false
			btns_to_clear = append(btns_to_clear, elevator_io.BT_HallUp)
		} else if e.orders[e.floor][elevator_io.BT_HallDown]{
			e.orders[e.floor][elevator_io.BT_HallDown] = false
			btns_to_clear = append(btns_to_clear, elevator_io.BT_HallDown)
		}
	}

	return btns_to_clear
}
