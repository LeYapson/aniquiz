package game

import (
    "testing"
    "time"
)

func TestRoomStateTransition(t *testing.T) {
    room := CreateRoom("test-room")
    
    // On vérifie l'état initial
    if room.State != StateLobby {
        t.Errorf("Attendu StateLobby, obtenu %s", room.State)
    }

    // On lance la boucle Run dans une goroutine pour qu'elle tourne en arrière-plan
    go room.Run()

    // Simulation du clic sur "Commencer la partie"
    room.Start <- true

    // On laisse un tout petit peu de temps au processeur pour traiter le channel
    time.Sleep(10 * time.Millisecond)

    if room.State != StatePlaying {
        t.Errorf("La room aurait dû passer en StatePlaying après le signal Start")
    }
}