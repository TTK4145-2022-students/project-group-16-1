package elevator_control

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type ClearRequestVariant int

const (
	CV_All ClearRequestVariant = iota
	CV_InDirn
)

type Elevator struct {
	floor     int
	dirn      Dirn
	requests  [N_FLOORS][N_BUTTONS]int
	behaviour ElevatorBehaviour

	config struct {
		clearRequestVariant ClearRequestVariant
		doorOpenDuration_s  float64 //isteden for double
	}
}

func eb_toString(ev ElevatorBehaviour) {

}

func elevator_print() {

}

func elevator_uninitialised() {
	return Elevator{
		floor = -1,
		dirn = D_Stop,
		behaviour = EB_Idle,
		config.clearRequestVariant = CV_All,
		config.doorOpenDuration_s = 3.0,
	}
}