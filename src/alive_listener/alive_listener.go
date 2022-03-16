package alive_listener

import (
	"Elevator-project/src/network/peers"
)

func AliveListener(
	peersUpdate <-chan peers.PeerUpdate,
	al_or_newElevDetected chan<- string,
	al_or_elevsLost chan<- []string,
	al_or_disconnected chan<- bool,
	oa_al_newElevDetected chan<- string,
	oa_al_elevsLost chan<- []string,
) {
	for {
		elevUpdate := <-peersUpdate
		if len(elevUpdate.New) != 0 {
			al_or_newElevDetected <- elevUpdate.New
			oa_al_newElevDetected <- elevUpdate.New
		}
		if len(elevUpdate.Lost) != 0 {
			al_or_elevsLost <- elevUpdate.Lost
			oa_al_elevsLost <- elevUpdate.Lost
		}
		if len(elevUpdate.Peers) == 0 {
			al_or_disconnected <- true
		}
	}
}
