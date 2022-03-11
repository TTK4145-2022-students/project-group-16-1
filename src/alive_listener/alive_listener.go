package alive_listener

import (
	"Elevator-project/src/network/peers"
)

func AliveListener(
	peersUpdate <-chan peers.PeerUpdate,
	orm_newElevDetected chan<- string,
	orm_elevsLost chan<- []string,
	orm_disconnected chan<- bool,
	oa_newElevDetected chan<- string,
	oa_elevsLost chan<- []string,
) {
	for {
		elevUpdate := <-peersUpdate
		if len(elevUpdate.New) != 0 {
			orm_newElevDetected <- elevUpdate.New
			oa_newElevDetected <- elevUpdate.New
		}
		if len(elevUpdate.Lost) != 0 {
			orm_elevsLost <- elevUpdate.Lost
			oa_elevsLost <- elevUpdate.Lost
		}
		if len(elevUpdate.Peers) == 0 {
			orm_disconnected <- true
		}
	}
}
