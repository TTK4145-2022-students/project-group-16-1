package alive_listener

import (
	"Elevator-project/src/network/peers"
	"fmt"
)

func AliveListener(
	peersUpdate 			<-chan peers.PeerUpdate,
	al_or_newElevDetected 	chan<- string,
	al_or_elevsLost 		chan<- []string,
	al_or_disconnected 		chan<- bool,
	al_oa_newElevDetected 	chan<- string,
	al_oa_elevsLost 		chan<- []string,
	id 						string,
) {
	for {
		elev_update := <-peersUpdate
		if len(elev_update.New) != 0 {
			al_or_newElevDetected <- elev_update.New
			al_oa_newElevDetected <- elev_update.New
		}
		if len(elev_update.Lost) != 0 {
			al_or_elevsLost <- elev_update.Lost
			al_oa_elevsLost <- elev_update.Lost
		}
		if contains(elev_update.Lost, id) {
			al_or_disconnected <- true
		}

		fmt.Printf("Peer update:\n")
		fmt.Printf("  Peers:    %q\n", elev_update.Peers)
		fmt.Printf("  New:      %q\n", elev_update.New)
		fmt.Printf("  Lost:     %q\n", elev_update.Lost)

	}
}

func contains(s []string, element string) bool {
	for _, str := range s {
		if str == element {
			return true
		}
	}
	return false
}
