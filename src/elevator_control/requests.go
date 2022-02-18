package elevator_control

type Action struct {
	dirn      Dirn
	behaviour ElevatorBehaviour
}

func requests_above(e Elevator) bool {
	for f := e.floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.requests[f][btn] != 0 {
				return true
			}
		}
	}
	return false
}

func requests_below(e Elevator) bool {
	for f := e.floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.requests[f][btn] != 0 {
				return true
			}
		}
	}
	return false
}

func requests_here(e Elevator) bool {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if e.requests[e.floor][btn] != 0 {
			return true
		}
	}
	return false
}

func requests_nextAction(e Elevator) Action {
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
		return action

	default:
		var action Action
		action.behaviour = EB_Idle
		action.dirn = D_Stop
		return action
	}
}

func requests_shouldStop(e Elevator) bool {
	switch e.dirn {
	case D_Down:
		return (e.requests[e.floor][B_HallDown] != 0) || (e.requests[e.floor][B_Cab] != 0) || !requests_below(e)
	case D_Up:
		return (e.requests[e.floor][B_HallUp] != 0) || (e.requests[e.floor][B_Cab] != 0) || !requests_above(e)
	case D_Stop:
		return true
	default:
		return true
	}
}

func requests_shouldClearImmediately(e Elevator, btn_floor int, btn_type Button) bool {
	switch e.config.clearRequestVariant {
	case CV_All:
		return e.floor == btn_floor
	case CV_InDirn:
		return (e.floor == btn_floor) &&
			((e.dirn == D_Up && btn_type == B_HallUp) ||
				(e.dirn == D_Down && btn_type == B_HallDown) ||
				e.dirn == D_Stop ||
				btn_type == B_Cab)
	default:
		return false
	}
}

func requests_clearAtCurrentFloor(e Elevator) Elevator {

	switch e.config.clearRequestVariant {
	case CV_All:
		//not relevant
	case CV_InDirn:
		e.requests[e.floor][B_Cab] = 0
		switch e.dirn {
		case D_Up:
			if !requests_above(e) && (e.requests[e.floor][B_HallUp] == 0) {
				e.requests[e.floor][B_HallDown] = 0
			}
			e.requests[e.floor][B_HallUp] = 0
		case D_Down:
			if !requests_below(e) && (e.requests[e.floor][B_HallDown] == 0) {
				e.requests[e.floor][B_HallUp] = 0
			}
			e.requests[e.floor][B_HallDown] = 0
		case D_Stop:
		default:
			e.requests[e.floor][B_HallUp] = 0
			e.requests[e.floor][B_HallDown] = 0
		}

	default:
	}
	return e
}
