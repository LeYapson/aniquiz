package game

import (
	"encoding/json"
	"fmt"
	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/LeYapson/aniquiz/internal/models"
	"log"
	"sync"
	"time"
)

var (
	// ActiveRooms stocke tous les salons en cours avec leur ID comme clé
	ActiveRooms = make(map[string]*Room)
	// RoomsMu protège l'accès à la map ActiveRooms pour éviter les crashs multijoueurs
	RoomsMu sync.Mutex
)

type RoomState string

const (
	StateLobby   RoomState = "LOBBY"
	StatePlaying RoomState = "PLAYING"
)

type Room struct {
	ID         string
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	State      RoomState
	// --- AJOUTS ICI ---
	CurrentTrack *models.Track // La musique que les gens doivent deviner
	Start        chan bool     // Canal pour recevoir l'ordre de démarrage
	IsPlaying    bool          // Le quiz a-t-il démarré ?
	// ------------------
	CurrentRound int
	MaxRounds    int
	RoundDuration int
	IsPrivate	 bool
	Password	 string
	CreatorID	 string

	Mu sync.Mutex
}

type RoomSummary struct {
    ID           string    `json:"id"`
    State        RoomState `json:"state"`
    PlayersCount int       `json:"players_count"`
    IsPrivate    bool      `json:"is_private"`
    MaxRounds    int       `json:"max_rounds"`
}

func GetPublicRooms() []RoomSummary {
    RoomsMu.Lock()
    defer RoomsMu.Unlock()

    var list []RoomSummary
    for id, room := range ActiveRooms {
        room.Mu.Lock()
        // On liste toutes les rooms, ou uniquement les publiques selon ton choix
        list = append(list, RoomSummary{
            ID:           id,
            State:        room.State,
            PlayersCount: len(room.Clients),
            IsPrivate:    room.IsPrivate,
            MaxRounds:    room.MaxRounds,
        })
        room.Mu.Unlock()
    }
    return list
}

func CreateRoom(id string, creatorID string) *Room {
	return &Room{
		ID:         id,
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Start:      make(chan bool),
		IsPlaying:  false, // On initialise à false à la place
		State:      StateLobby,
		CurrentRound: 0,
		MaxRounds: 5, // Par exemple, on peut faire 5 rounds par partie
		RoundDuration: 20, // Durée de chaque round en secondes
		Password: "",
		IsPrivate: false,
		CreatorID: creatorID,
	}
}

// Run est le moteur du salon : il tourne en boucle pour gérer les événements
func (r *Room) Run() {
	for {
		select {
		case client := <-r.Register:
			r.Clients[client] = true

			// 1. Liste des joueurs pour tout le monde (incluant le nouveau)
			go r.BroadcastPlayerList()

			// 2. Envoyer l'état actuel (LOBBY ou PLAYING) au nouveau venu
			msgState, _ := json.Marshal(map[string]interface{}{
				"type":    "GAME_STATE",
				"payload": r.State,
			})
			client.Send <- msgState

			// 3. Si une partie est en cours, on lui envoie la musique actuelle
			if r.State == StatePlaying && r.CurrentTrack != nil {
				msgTrack, _ := json.Marshal(map[string]interface{}{
					"type": "NewQuestion", // Assure-toi que c'est bien ce type que ton Front attend
					"payload": map[string]interface{}{
						"audio_url": r.CurrentTrack.AudioURL,
						"room_id":   r.ID,
					},
				})
				client.Send <- msgTrack
			}

		case <-r.Start:
			if r.State != StatePlaying {
				r.State = StatePlaying
				r.broadcastGameState() // On prévient le Front
				go r.nextRound()

				// Ici, on lancera plus tard la fonction qui choisit une musique
				log.Println("La partie commence dans le salon:", r.ID)
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

// Fonction pour envoyer l'état au Front
func (r *Room) broadcastGameState() {
	msg := map[string]interface{}{
		"type":    "GAME_STATE",
		"payload": r.State,
	}
	data, _ := json.Marshal(msg)
	for c := range r.Clients {
		c.Send <- data
	}
}

// BroadcastPlayerList envoie la liste des joueurs à tous les clients du salon
func (r *Room) BroadcastPlayerList() {
	var players []models.PlayerInfo
	for c := range r.Clients {
		players = append(players, models.PlayerInfo{
			ID:       c.ID,
			Username: c.Username,
			Score:    c.Score, // On peut ajouter la logique de score plus tard
		})
	}

	msg := models.WSMessage{
		Type:    "PLAYER_LIST",
		Payload: players,
	}

	data, _ := json.Marshal(msg)
	r.Broadcast <- data
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

	// Après un délai, on peut lancer la prochaine question
	time.Sleep(500 * time.Millisecond)
	go r.nextRound()

}

func (r *Room) CheckAnswer(client *Client, answer string) {
	r.Mu.Lock()
	if !r.IsPlaying || r.CurrentTrack == nil {
		r.Mu.Unlock()
		return
	}
	track := r.CurrentTrack
	r.Mu.Unlock()

	result := VerifyAnswer(answer, track)

	if result.Points > 0 {
		// 1- mise a jour du score du client
		client.Score += result.Points

		//2- annonce du gain de points
		msg := map[string]interface{}{
			"type": "PLAYER_GUESS",
			"payload": map[string]interface{}{
				"username": client.Username,
				"message":  "a trouvé l'anime !", // On reste discret sur le nom
			},
		}
		data, _ := json.Marshal(msg)
		r.Broadcast <- data

		//3- Renvoyer la liste des joueurs mise à jour avec les scores
		go r.BroadcastPlayerList()
	}
}

func (r *Room) nextRound() {
	// 1. Chercher une musique aléatoire via ton package database
	r.Mu.Lock()
	if r.CurrentRound >= r.MaxRounds {
		r.Mu.Unlock()
		r.finishGame()
		return
	}
	r.CurrentRound++
	// On récupere la durée sous forme de variable locale pour le timer
	duration := r.RoundDuration
	r.Mu.Unlock()
	
	track, err := database.GetRandomTrack()
	if err != nil {
		log.Printf("Erreur récup musique: %v", err)
		return
	}
	r.Mu.Lock()
	r.CurrentTrack = track
	r.IsPlaying = true
	r.Mu.Unlock()

	log.Printf("Nouveau round dans %s : %s", r.ID, track.AnimeName)

	// 2. Préparer le message pour le Front
	// On n'envoie QUE l'URL, pas la réponse !
	msg := map[string]interface{}{
		"type": "NewQuestion",
		"payload": map[string]interface{}{
			"audio_url": track.AudioURL,
			"room_id":   r.ID,
			"duration":  duration, // On peut aussi envoyer la durée du round
		},
	}

	// 3. Envoyer à tout le monde
	data, _ := json.Marshal(msg)

	r.Broadcast <- data

	// 4. Lancer un timer pour la fin du round
	go func() {
		time.Sleep(time.Duration(duration) * time.Second)
		r.EndRound("Temps écoulé !")
	}()
}

func (r *Room) finishGame() {
    r.Mu.Lock()
    r.State = StateLobby
    r.IsPlaying = false
    r.CurrentRound = 0
    r.CurrentTrack = nil
    r.Mu.Unlock()

    // 1. On prévient le Front que c'est fini
    msg := map[string]interface{}{
        "type": "GAME_OVER",
        "payload": map[string]interface{}{
            "message": "La partie est terminée !",
        },
    }
    data, _ := json.Marshal(msg)
    r.Broadcast <- data

    // 2. On remet les scores de tout le monde à 0
    for client := range r.Clients {
        client.Score = 0
    }

    // 3. On renvoie les états mis à jour au Front (Lobby + Scores à 0)
    r.broadcastGameState()
    go r.BroadcastPlayerList()
}
