package allive_pinger_listener

import (
	"Elevator-project/src/network/peers"
)

func AllivePingerListener(
	peersUpdate <-chan peers.PeerUpdate,
	newElevDetected chan<- string,
	elevsLost chan<- []string,
	disconnected chan<- bool,
) {
	for {
		select {
		case elevUpdate := <-peersUpdate:
			if len(elevUpdate.New) != 0 {
				newElevDetected <- elevUpdate.New
			}
			if len(elevUpdate.Lost) != 0 {
				elevsLost <- elevUpdate.Lost
			}
			if len(elevUpdate.Peers) == 0 {
				disconnected <- true
			}
		}

	}
}
