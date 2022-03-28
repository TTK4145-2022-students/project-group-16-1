package order_assigner

import (
	. "Elevator-project/src/constants"
	"Elevator-project/src/elevator_control"
	"Elevator-project/src/order_redundancy"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

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
	or_oa_confirmedOrders <-chan order_redundancy.ConfirmedOrders,
	net_oa_elevatorState <-chan elevator_control.ElevatorMsg,
	ec_oa_elevator <-chan elevator_control.ElevatorMsg,
	al_oa_newElevDetected <-chan string,
	al_oa_elevsLost <-chan []string,
	oa_ec_assignedOrders chan<- [N_FLOORS][N_BTN_TYPES]bool,
	id string) {

	confirmed_orders := order_redundancy.ConfirmedOrders{}
	elevator_states := make(map[string]elevator_control.ElevatorMsg)

	elevator_states[id] = elevator_control.ElevatorMsg{}

	//Init -> get first local elevator state
	<-ec_oa_elevator

	for {
		select {
		case confirmed_orders = <-or_oa_confirmedOrders:

			local_assigned_orders, err := assign(confirmed_orders, elevator_states, id)
			if !err {
				oa_ec_assignedOrders <- local_assigned_orders
			}

		case new_elev := <-al_oa_newElevDetected:
			elevator_states[new_elev] = elevator_control.ElevatorMsg{}

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

		case state := <-net_oa_elevatorState:
			if prev_state, ok := elevator_states[state.Id]; ok {

				if prev_state != state {

					elevator_states[state.Id] = state
					local_assigned_orders, err := assign(confirmed_orders, elevator_states, id)
					if !err {
						oa_ec_assignedOrders <- local_assigned_orders
					}
				}
			}
		case state := <-ec_oa_elevator:
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

func generateJson(orders order_redundancy.ConfirmedOrders, states map[string]elevator_control.ElevatorMsg) []byte {
	var msg JsonProxy
	msg.HallRequests = orders.HallCalls
	msg.States = make(map[string]JsonIdState)
	for id, elev_state := range states {
		if elev_state.Available {
			if _, ok := orders.CabCalls[id]; ok {
				msg.States[id] = JsonIdState{elev_state.Behaviour, elev_state.Floor, elev_state.Dirn, orders.CabCalls[id]}
			}
		}
	}
	json_msg, _ := json.Marshal(msg)
	return json_msg
}

func assign(orders order_redundancy.ConfirmedOrders,
	states map[string]elevator_control.ElevatorMsg, id string) ([N_FLOORS][N_BTN_TYPES]bool, bool) {
	json_msg := generateJson(orders, states)
	if !states[id].Available {
		return [N_FLOORS][N_BTN_TYPES]bool{}, true
	}
	assigned_orders, err := exec.Command("./src/order_assigner/hall_request_assigner", "--input", string(json_msg), "--includeCab").Output()
	if err != nil {
		fmt.Println("order assigner failed -don't worry, be happy")
		return [N_FLOORS][N_BTN_TYPES]bool{}, true
	}
	var msg map[string][N_FLOORS][N_BTN_TYPES]bool
	err = json.Unmarshal(assigned_orders, &msg)
	if err != nil {
		log.Fatal(err)
	}
	return msg[id], false
}
