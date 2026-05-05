package game

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/LeYapson/aniquiz/internal/models"
)

var (
	// ActiveRooms stocke tous les salons en cours avec leur ID comme clé
	ActiveRooms = make(map[string]*Room)
	// RoomsMu protège l'accès à la map ActiveRooms pour éviter les crashs multijoueurs
	RoomsMu     sync.Mutex
)

type RoomState string

const (
    StateLobby   RoomState = "LOBBY"
    StatePlaying RoomState = "PLAYING"
)

type Room struct {
	ID           string
	Clients      map[*Client]bool
	Broadcast    chan []byte
	Register     chan *Client
	Unregister   chan *Client
	State        RoomState
	// --- AJOUTS ICI ---
	CurrentTrack *models.Track // La musique que les gens doivent deviner
	IsPlaying    bool          // Le quiz a-t-il démarré ?
	// ------------------

	Mu           sync.Mutex
}

func CreateRoom(id string) *Room {
	return &Room{
		ID:         id,
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
        IsPlaying:  false, // On initialise à false à la place
	}
}

// Run est le moteur du salon : il tourne en boucle pour gérer les événements
func (r *Room) Run() {
	for {
		select {
		case client := <-r.Register:
			r.Clients[client] = true
            go r.BroadcastPlayerList() // On envoie la liste des joueurs à tout le monde

            // SYNCHRONISATION :
            r.Mu.Lock()
            if r.IsPlaying && r.CurrentTrack != nil {
                syncMsg := models.WSMessage{
                    Type: "NEW_TRACK",
                    Payload: map[string]interface{}{
                        "audio_url": r.CurrentTrack.AudioURL,
                        "duration":  20, // Idéalement, calculer le temps restant
                    },
                }
                data, _ := json.Marshal(syncMsg)
                client.Send <- data
            }
            r.Mu.Unlock()
            // TEST : Si c'est le premier joueur, on lance une musique après 2 secondes
			if len(r.Clients) == 1 {
				go func() {
					time.Sleep(2 * time.Second)
					r.StartNextRound()
				}()
			}

		case client := <-r.Unregister:
			if _, ok := r.Clients[client]; ok {
				delete(r.Clients, client)
				close(client.Send)
                // Si le salon est vide, on arrête le quiz
                if len(r.Clients) == 0 {
                    r.Mu.Lock()
                    r.IsPlaying = false
                    r.Mu.Unlock()
                    fmt.Printf("Salon %s vide, mise en pause.\n", r.ID)
                } else {
                    go r.BroadcastPlayerList() // On met à jour la liste des joueurs
                }
			}
		case message := <-r.Broadcast:
			for client := range r.Clients {
				client.Send <- message
			}
		}
	}
}

// BroadcastPlayerList envoie la liste des joueurs à tous les clients du salon
func (r *Room) BroadcastPlayerList() {
    var players []models.PlayerInfo
    for c:= range r.Clients {
        players = append(players, models.PlayerInfo{
            ID: c.ID,
            Username: c.Username,
            Score: c.Score, // On peut ajouter la logique de score plus tard
        })
    }

    msg := models.WSMessage{
        Type: "PLAYER_LIST",
        Payload: players,
    }

    data, _ := json.Marshal(msg)
    r.Broadcast <- data
}

func(r *Room) StartNextRound() {
    //1. on récupere une musique aleatoire via notre package database
    track, err := database.GetRandomTrack()
    if err != nil {
        fmt.Printf("Erreur récup musique: %v\n", err)
        return
    }
    
    r.Mu.Lock()
	r.CurrentTrack = track
	r.IsPlaying = true
	r.Mu.Unlock()

    //2. Préparer le message pour les joueurs (on cache le nom de l'anime !)
    msg := models.WSMessage{
        Type: "NEW_TRACK",
        Payload: map[string]interface{}{
            "audio_url": track.AudioURL,
            "duration": 20, //on leur dit qu'ils ont 20 secondes pour répondre
        },
    }

    data, _ := json.Marshal(msg)

    //3. Envoyer à tous les joueurs du salon
    r.Broadcast <- data

    go func() {
        time.Sleep(20 * time.Second)
        r.EndRound("Temps écoulé !")
    }()
}

func (r *Room) EndRound(reason string) {
    r.Mu.Lock()
    if !r.IsPlaying {
        r.Mu.Unlock()
        return
    }

    // On récupere la réponse correcte pour l'anime
    reveal := r.CurrentTrack.AnimeName
    r.IsPlaying = false
    r.Mu.Unlock()

    msg := models.WSMessage{
        Type: "ROUND_ENDED",
        Payload: map[string]interface{}{
            "reason": reason,
            "answer": reveal,
        },
    }

    data, _ := json.Marshal(msg)
    r.Broadcast <- data

    // optionnel : Relancer un round  automatiquement après quelques secondes
    time.Sleep(5 * time.Second)
    r.StartNextRound()
}

func (r *Room) CheckAnswer(client *Client, answer string) {
    r.Mu.Lock()
    if !r.IsPlaying || r.CurrentTrack == nil {
        r.Mu.Unlock()
        return
    }
    r.Mu.Unlock()

    result := VerifyAnswer(answer, r.CurrentTrack)

    if result.Points > 0 {
        // 1- mise a jour du score du client
        client.Score += result.Points

        //2- annonce du gain de points
        msg := models.WSMessage{
            Type: "PLAYER_GUESS",
            Payload: map[string]interface{}{
                "username": client.Username,
                "points": result.Points,
                "message": result.Message,
            },
        }
        data, _ := json.Marshal(msg)
        r.Broadcast <- data

        //3- Renvoyer la liste des joueurs mise à jour avec les scores
        go r.BroadcastPlayerList()

        //4- Si c'est la réponse parfaite, on termine le round
        if result.IsCorrect {
            // On laisse un petit délai pour que les autres tentent le nom de la musique ?
			// Ou on finit direct :
			// go r.EndRound("Bonne réponse trouvée !")
        }
    }
}
