package order_assigner

import (
	"Elevator-project/src/elevator_control"
	"Elevator-project/src/order_redundancy_manager"
	"encoding/json"
	"log"
	"os/exec"
)

const N_FLOORS = 4 //REMOVE THIS
const N_BTN_TYPES = 3

type JsonIdState struct {
	Behaviour   string          `json:"behaviour"`
	Floor       int             `json:"floor"`
	Direction   string          `json:"direction"`
	CabRequests *[N_FLOORS]bool `json:"cabRequests"`
}

type JsonProxy struct {
	HallRequests [N_FLOORS][2]bool      `json:"hallRequests"`
	States       map[string]JsonIdState `json:"states"`
}

func OrderAssigner(
	or_oa_confirmedOrders <-chan order_redundancy_manager.ConfirmedOrders,
	net_elevatorState <-chan elevator_control.ElevatorState,
	ec_oa_localElevatorState <-chan elevator_control.ElevatorState,
	al_oa_newElevDetected <-chan string,
	al_oa_elevsLost <-chan []string,
	oa_ec_assignedOrders chan<- [N_FLOORS][N_BTN_TYPES]bool,
	id string) {

	confirmed_orders := order_redundancy_manager.ConfirmedOrders{}
	elevator_states := make(map[string]elevator_control.ElevatorState)

	elevator_states[id] = elevator_control.ElevatorState{}

	//Init -> get first local elevator state
	<-ec_oa_localElevatorState

	for {
		select {
		case confirmed_orders = <-or_oa_confirmedOrders:

			local_assigned_orders, err := assign(confirmed_orders, elevator_states, id)
			if !err {
				oa_ec_assignedOrders <- local_assigned_orders
			}

		case new_elev := <-al_oa_newElevDetected:
			elevator_states[new_elev] = elevator_control.ElevatorState{}

		case lost_elev := <-al_oa_elevsLost:
			for i := range lost_elev {
				if lost_elev[i] == id {
					break
				}
				delete(elevator_states, lost_elev[i])
			}
			local_assigned_orders, err := assign(confirmed_orders, elevator_states, id)
			if !err {
				oa_ec_assignedOrders <- local_assigned_orders
			}

		case state := <-net_elevatorState:
			if prev_state, ok := elevator_states[state.Id]; ok {

				if prev_state != state {

					elevator_states[state.Id] = state
					local_assigned_orders, err := assign(confirmed_orders, elevator_states, id)
					if !err {
						oa_ec_assignedOrders <- local_assigned_orders
					}
				}
			}
		case state := <-ec_oa_localElevatorState:
			if prev_state, ok := elevator_states[state.Id]; ok {

				if prev_state != state {

					elevator_states[state.Id] = state
					local_assigned_orders, err := assign(confirmed_orders, elevator_states, id)
					if !err {
						oa_ec_assignedOrders <- local_assigned_orders
					}
				}
			}
		}
	}

}

func generateJson(orders order_redundancy_manager.ConfirmedOrders, states map[string]elevator_control.ElevatorState) []byte {
	var msg JsonProxy
	msg.HallRequests = orders.HallCalls
	msg.States = make(map[string]JsonIdState)
	for id, elev_state := range states {
		switch elev_state.Behaviour {
		case elevator_control.EB_DoorOpen.String(),
			elevator_control.EB_Idle.String(),
			elevator_control.EB_Moving.String():
			if _, ok := orders.CabCalls[id]; ok {
				msg.States[id] = JsonIdState{elev_state.Behaviour, elev_state.Floor, elev_state.Dirn, orders.CabCalls[id]}
			}
		}
	}
	json_msg, _ := json.Marshal(msg)
	return json_msg
}

func assign(orders order_redundancy_manager.ConfirmedOrders,
	states map[string]elevator_control.ElevatorState, id string) ([N_FLOORS][N_BTN_TYPES]bool, bool) {
	json_msg := generateJson(orders, states)

	assigned_orders, err := exec.Command("./src/order_assigner/hall_request_assigner", "--input", string(json_msg), "--includeCab").Output()
	if err != nil {
		log.Fatal(err)
	}
	var msg map[string][N_FLOORS][N_BTN_TYPES]bool
	err = json.Unmarshal(assigned_orders, &msg)
	if err != nil {
		log.Fatal(err)
	}
	return msg[id], false
}
