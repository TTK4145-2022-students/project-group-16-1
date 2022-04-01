A peer-to-peer distributed system.

Initialise new elevator in terminal:
go run main.go -id="<id>"

Module overview:

* al    - alive_listener:       Listens to pings from other elevators on network, maintains list of alive elevators.
* ec    - elevator_control:     Controls the elevator to take all asigned orders (input).
* eio   - elevator_io:          Polls hardware.
* net   - network:              Sends and recieves orders, elevator states and alive pings over network.
* oa    - order_assigner:       Recieves orders and states from all elevators and "optimally" assigns these between elevators.
* or    - order_redundancy:     Handles state transitions of orders. Order states are UNKNOWN, NONE, UNCONFIRMED and CONFIRMED.