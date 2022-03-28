package elevator_control

import (
	. "Elevator-project/src/constants"
	"Elevator-project/src/elevio"
)

type Action struct {
	dirn      elevio.MotorDirection
	behaviour ElevatorBehaviour
}

func requests_above(e *Elevator) bool {
	for f := e.floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BTN_TYPES; btn++ {
			if e.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_below(e *Elevator) bool {
	for f := 0; f < e.floor; f++ {
		for btn := 0; btn < N_BTN_TYPES; btn++ {
			if e.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_here(e *Elevator) bool {
	for btn := 0; btn < N_BTN_TYPES; btn++ {
		if e.requests[e.floor][btn] {
			return true
		}
	}
	return false
}

func requests_nextAction(e *Elevator) Action {
	switch e.dirn {
	case elevio.MD_Up:
		var action Action

		if requests_above(e) {
			action.behaviour = EB_Moving
			action.dirn = elevio.MD_Up
		} else if requests_here(e) {
			action.behaviour = EB_DoorOpen
			action.dirn = elevio.MD_Down
		} else if requests_below(e) {
			action.behaviour = EB_Moving
			action.dirn = elevio.MD_Down
		} else {
			action.behaviour = EB_Idle
			action.dirn = elevio.MD_Stop
		}

		return action

	case elevio.MD_Down:
		var action Action

		if requests_below(e) {
			action.behaviour = EB_Moving
			action.dirn = elevio.MD_Down
		} else if requests_here(e) {
			action.behaviour = EB_DoorOpen
			action.dirn = elevio.MD_Up
		} else if requests_above(e) {
			action.behaviour = EB_Moving
			action.dirn = elevio.MD_Up
		} else {
			action.behaviour = EB_Idle
			action.dirn = elevio.MD_Stop
		}

		return action
	case elevio.MD_Stop: // there should only be one request in the Stop case. Checking up or down first is arbitrary.
		var action Action

		if requests_here(e) {
			action.behaviour = EB_DoorOpen
			action.dirn = elevio.MD_Stop
		} else if requests_below(e) {
			action.behaviour = EB_Moving
			action.dirn = elevio.MD_Down
		} else if requests_above(e) {
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

func requests_shouldStop(e *Elevator) bool {
	switch e.dirn {
	case elevio.MD_Down:
		return (e.requests[e.floor][elevio.BT_HallDown]) || (e.requests[e.floor][elevio.BT_Cab]) || !requests_below(e)
	case elevio.MD_Up:
		return (e.requests[e.floor][elevio.BT_HallUp]) || (e.requests[e.floor][elevio.BT_Cab]) || !requests_above(e)
	case elevio.MD_Stop:
		return true
	default:
		return true
	}
}

func requests_shouldClearImmediately(e *Elevator) []elevio.ButtonType {
	should_clear_btns := []elevio.ButtonType{}

	if e.dirn == elevio.MD_Up {
		if e.requests[e.floor][elevio.BT_HallUp] == true {
			e.requests[e.floor][elevio.BT_HallUp] = false
			should_clear_btns = append(should_clear_btns, elevio.BT_HallUp)
		}
	} else if e.dirn == elevio.MD_Down {
		if e.requests[e.floor][elevio.BT_HallDown] == true {
			e.requests[e.floor][elevio.BT_HallDown] = false
			should_clear_btns = append(should_clear_btns, elevio.BT_HallDown)
		}
	} else {
		if e.requests[e.floor][elevio.BT_HallUp] == true {
			e.requests[e.floor][elevio.BT_HallUp] = false
			should_clear_btns = append(should_clear_btns, elevio.BT_HallUp)
			e.dirn = elevio.MD_Up
		} else if e.requests[e.floor][elevio.BT_HallDown] == true {
			e.requests[e.floor][elevio.BT_HallDown] = false
			should_clear_btns = append(should_clear_btns, elevio.BT_HallDown)
			e.dirn = elevio.MD_Down
		}
	}
	if e.requests[e.floor][elevio.BT_Cab] == true {
		e.requests[e.floor][elevio.BT_Cab] = false
		should_clear_btns = append(should_clear_btns, elevio.BT_Cab)
	}

	return should_clear_btns
}

func requests_clearAtCurrentFloor(e *Elevator) []elevio.ButtonType {
	should_clear_btns := []elevio.ButtonType{}
	if e.requests[e.floor][elevio.BT_Cab] == true {
		e.requests[e.floor][elevio.BT_Cab] = false
		should_clear_btns = append(should_clear_btns, elevio.BT_Cab)
	}
	switch e.dirn {
	case elevio.MD_Up:
		if !requests_above(e) && (!e.requests[e.floor][elevio.BT_HallUp]) {
			if e.requests[e.floor][elevio.BT_HallDown] == true {
				e.requests[e.floor][elevio.BT_HallDown] = false
				should_clear_btns = append(should_clear_btns, elevio.BT_HallDown)
			}
		}
		if e.requests[e.floor][elevio.BT_HallUp] == true {
			e.requests[e.floor][elevio.BT_HallUp] = false
			should_clear_btns = append(should_clear_btns, elevio.BT_HallUp)
		}

	case elevio.MD_Down:
		if !requests_below(e) && (!e.requests[e.floor][elevio.BT_HallDown]) {

			if e.requests[e.floor][elevio.BT_HallUp] == true {
				e.requests[e.floor][elevio.BT_HallUp] = false
				should_clear_btns = append(should_clear_btns, elevio.BT_HallUp)
			}
		}
		if e.requests[e.floor][elevio.BT_HallDown] == true {
			e.requests[e.floor][elevio.BT_HallDown] = false
			should_clear_btns = append(should_clear_btns, elevio.BT_HallDown)
		}
	case elevio.MD_Stop:
		if e.requests[e.floor][elevio.BT_HallUp] == true {
			e.requests[e.floor][elevio.BT_HallUp] = false
			should_clear_btns = append(should_clear_btns, elevio.BT_HallUp)
		} else if e.requests[e.floor][elevio.BT_HallDown] == true {
			e.requests[e.floor][elevio.BT_HallDown] = false
			should_clear_btns = append(should_clear_btns, elevio.BT_HallDown)
		}
	}

	return should_clear_btns
}
