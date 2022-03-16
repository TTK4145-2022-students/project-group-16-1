package elevator_control

import "fmt"

type Action struct {
	dirn      Dirn
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
	case D_Up:
		var action Action

		if requests_above(e) {
			action.behaviour = EB_Moving
			action.dirn = D_Up
		} else if requests_here(e) {
			action.behaviour = EB_DoorOpen
			action.dirn = D_Down
		} else if requests_below(e) {
			action.behaviour = EB_Moving
			action.dirn = D_Down
		} else {
			action.behaviour = EB_Idle
			action.dirn = D_Stop
		}

		return action

	case D_Down:
		var action Action

		if requests_below(e) {
			action.behaviour = EB_Moving
			action.dirn = D_Down
		} else if requests_here(e) {
			action.behaviour = EB_DoorOpen
			action.dirn = D_Up
		} else if requests_above(e) {
			action.behaviour = EB_Moving
			action.dirn = D_Up
		} else {
			action.behaviour = EB_Idle
			action.dirn = D_Stop
		}

		return action
	case D_Stop: // there should only be one request in the Stop case. Checking up or down first is arbitrary.
		var action Action

		if requests_here(e) {
			action.behaviour = EB_DoorOpen
			action.dirn = D_Stop
		} else if requests_below(e) {
			action.behaviour = EB_Moving
			action.dirn = D_Down
		} else if requests_above(e) {
			action.behaviour = EB_Moving
			action.dirn = D_Up
		} else {
			action.behaviour = EB_Idle
			action.dirn = D_Stop
		}
		fmt.Printf("%+v", action)
		return action

	default:
		var action Action
		action.behaviour = EB_Idle
		action.dirn = D_Stop
		return action
	}
}

func requests_shouldStop(e *Elevator) bool {
	switch e.dirn {
	case D_Down:
		return (e.requests[e.floor][B_HallDown]) || (e.requests[e.floor][B_Cab]) || !requests_below(e)
	case D_Up:
		return (e.requests[e.floor][B_HallUp]) || (e.requests[e.floor][B_Cab]) || !requests_above(e)
	case D_Stop:
		return true
	default:
		return true
	}
}

func requests_shouldClearImmediately(e *Elevator) {
	switch e.config.clearRequestVariant {
	case CV_All:
		//Not needed yet
	case CV_InDirn:
		if e.dirn == D_Up {
			e.requests[e.floor][B_HallUp] = false
		} else if e.dirn == D_Down {
			e.requests[e.floor][B_HallDown] = false
		} else {
			e.requests[e.floor][B_HallUp] = false
			e.requests[e.floor][B_HallDown] = false
		}
		e.requests[e.floor][B_Cab] = false
	default:
	}
}

func requests_clearAtCurrentFloor(e *Elevator) {

	switch e.config.clearRequestVariant {
	case CV_All:
		//not relevant
	case CV_InDirn:
		e.requests[e.floor][B_Cab] = false
		switch e.dirn {
		case D_Up:
			if !requests_above(e) && (!e.requests[e.floor][B_HallUp]) {
				e.requests[e.floor][B_HallDown] = false
			}
			e.requests[e.floor][B_HallUp] = false
		case D_Down:
			if !requests_below(e) && (!e.requests[e.floor][B_HallDown]) {
				e.requests[e.floor][B_HallUp] = false
			}
			e.requests[e.floor][B_HallDown] = false
		case D_Stop:
			e.requests[e.floor][B_HallUp] = false
			e.requests[e.floor][B_HallDown] = false
		default:
			e.requests[e.floor][B_HallUp] = false
			e.requests[e.floor][B_HallDown] = false
		}

	default:
	}
}
