package order_assigner

import (
	"Elevator-project/src/order_redundancy_manager"
	"Elevator-project/src/order_state_pub"
	"encoding/json"
	"log"
	"os/exec"
)

const N_FLOORS = 4 //REMOVE THIS
const N_BUTTONS = 3

type JsonIdState struct {
	Behaviour   string         `json:"behaviour"`
	Floor       int            `json:"floor"`
	Direction   string         `json:"direction"`
	CabRequests [N_FLOORS]bool `json:"cabRequests"`
}

type JsonProxy struct {
	HallRequests [N_FLOORS][2]bool      `json:"hallRequests"`
	States       map[string]JsonIdState `json:"states"`
}

func OrderAssigner(assigner_assignedOrders chan [N_FLOORS][N_BUTTONS]bool,
	orm_confirmedOrders chan order_redundancy_manager.ConfirmedOrders,
	osp_elevatorState chan map[string]order_state_pub.ElevatorState, id string) {

	var confirmed_orders order_redundancy_manager.ConfirmedOrders
	elevator_states := make(map[string]order_state_pub.ElevatorState)

	confirmed_orders = <-orm_confirmedOrders
	elevator_states = <-osp_elevatorState
	local_assigned_orders := assign(confirmed_orders, elevator_states, id)
	assigner_assignedOrders <- local_assigned_orders
	// elevator_states["one"] = order_state_pub.ElevatorState{"moving", 2, "up"}
	// elevator_states["two"] = order_state_pub.ElevatorState{"idle", 0, "stop"}

	for {
		select {
		case confirmed_orders = <-orm_confirmedOrders:
			local_assigned_orders := assign(confirmed_orders, elevator_states, id)
			assigner_assignedOrders <- local_assigned_orders
			print("conf orders recieved")
		case elevator_states = <-osp_elevatorState:
			local_assigned_orders := assign(confirmed_orders, elevator_states, id)
			assigner_assignedOrders <- local_assigned_orders
			print("state recieved")

		default:

		}
	}

}

func generateJson(orders order_redundancy_manager.ConfirmedOrders, states map[string]order_state_pub.ElevatorState) []byte {
	var msg JsonProxy
	msg.HallRequests = orders.HallCalls
	msg.States = make(map[string]JsonIdState)
	for id, elev_state := range states {
		if _, ok := orders.CabCalls[id]; ok {
			msg.States[id] = JsonIdState{elev_state.Behaviour, elev_state.Floor, elev_state.Dirn, orders.CabCalls[id]}
		}
	}
	json_msg, _ := json.Marshal(msg)
	return json_msg
}

func assign(orders order_redundancy_manager.ConfirmedOrders,
	states map[string]order_state_pub.ElevatorState, id string) [N_FLOORS][N_BUTTONS]bool {

	json_msg := generateJson(orders, states)
	assigned_orders, err := exec.Command("./src/order_assigner/hall_request_assigner", "--input", string(json_msg), "--includeCab").Output()
	if err != nil {
		log.Fatal(err)
	}
	var msg map[string][N_FLOORS][N_BUTTONS]bool
	err = json.Unmarshal(assigned_orders, &msg)
	if err != nil {
		log.Fatal(err)
	}
	return msg[id]
}
