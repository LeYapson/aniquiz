package game

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/LeYapson/aniquiz/internal/models"
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
	ID            string
	Clients       map[*Client]bool
	Broadcast     chan []byte
	Register      chan *Client
	Unregister    chan *Client
	State         RoomState
	CurrentTrack  *models.Track
	Start         chan bool
	IsPlaying     bool
	CurrentRound  int
	MaxRounds     int
	RoundDuration int
	IsPrivate     bool
	Password      string
	CreatorID     string
	HasAnswered   map[string]bool
	// Filtres de piste
	FilterType string // "OP", "ED", "" (tout)
	MinYear    int    // 0 = pas de filtre
	MaxYear    int    // 0 = pas de filtre

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
		ID:            id,
		Clients:       make(map[*Client]bool),
		Broadcast:     make(chan []byte),
		Register:      make(chan *Client),
		Unregister:    make(chan *Client),
		Start:         make(chan bool),
		IsPlaying:     false, // On initialise à false à la place
		State:         StateLobby,
		CurrentRound:  0,
		MaxRounds:     5,  // Par exemple, on peut faire 5 rounds par partie
		RoundDuration: 20, // Durée de chaque round en secondes
		Password:      "",
		IsPrivate:     false,
		CreatorID:     creatorID,
		HasAnswered:   make(map[string]bool),
	}
}

// Run est le moteur du salon : il tourne en boucle pour gérer les événements
func (r *Room) Run() {
	for {
		select {
		case client := <-r.Register:
			// Joueur arrivant en cours de partie → spectateur
			if r.State == StatePlaying {
				client.IsSpectator = true
			}
			r.Clients[client] = true

			// 1. Notifier le client de son statut
			statusMsg, _ := json.Marshal(map[string]interface{}{
				"type":    "SPECTATOR_STATUS",
				"payload": client.IsSpectator,
			})
			client.Send <- statusMsg

			// 2. Liste des joueurs pour tout le monde (incluant le nouveau)
			go r.BroadcastPlayerList()

			// 3. Envoyer l'état actuel (LOBBY ou PLAYING) au nouveau venu
			msgState, _ := json.Marshal(map[string]interface{}{
				"type":    "GAME_STATE",
				"payload": r.State,
			})
			client.Send <- msgState

			// 4. Si une partie est en cours, envoyer la musique actuelle au spectateur
			if r.State == StatePlaying && r.CurrentTrack != nil {
				msgTrack, _ := json.Marshal(map[string]interface{}{
					"type": "NewQuestion",
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

// BroadcastPlayerList envoie la liste des joueurs et spectateurs à tous les clients du salon.
func (r *Room) BroadcastPlayerList() {
	var players []models.PlayerInfo
	spectatorCount := 0
	for c := range r.Clients {
		if c.IsSpectator {
			spectatorCount++
			continue
		}
		players = append(players, models.PlayerInfo{
			ID:       c.ID,
			Username: c.Username,
			Score:    c.Score,
		})
	}

	data, _ := json.Marshal(map[string]interface{}{
		"type": "PLAYER_LIST",
		"payload": map[string]interface{}{
			"players":         players,
			"spectator_count": spectatorCount,
		},
	})
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
	time.Sleep(10 * time.Second)
	go r.nextRound()

}

func (r *Room) CheckAnswer(client *Client, answer string) {
	r.Mu.Lock()
	if !r.IsPlaying || r.CurrentTrack == nil {
		r.Mu.Unlock()
		return
	}

	if r.HasAnswered[client.ID] {
		r.Mu.Unlock()
		return // Le joueur a déjà validé ce round, on ignore son message
	}

	track := r.CurrentTrack
	r.Mu.Unlock()

	result := VerifyAnswer(answer, track)

	if result.Points > 0 {
		r.Mu.Lock()
		r.HasAnswered[client.ID] = true // Marquer que ce client a répondu pour ce round
		r.Mu.Unlock()
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

	r.HasAnswered = make(map[string]bool) // Réinitialiser les réponses pour le nouveau round
	// On récupere la durée sous forme de variable locale pour le timer
	duration := r.RoundDuration
	r.Mu.Unlock()

	r.Mu.Lock()
	filters := models.TrackFilters{
		TrackType: r.FilterType,
		MinYear:   r.MinYear,
		MaxYear:   r.MaxYear,
	}
	r.Mu.Unlock()

	track, err := database.GetRandomTrackFiltered(filters)
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

	// 1. Distribuer l'XP avant de réinitialiser les scores
	r.grantXP()

	// Lever le statut spectateur pour tous : la prochaine partie est ouverte à tous
	for c := range r.Clients {
		if c.IsSpectator {
			c.IsSpectator = false
			msg, _ := json.Marshal(map[string]interface{}{
				"type":    "SPECTATOR_STATUS",
				"payload": false,
			})
			select {
			case c.Send <- msg:
			default:
			}
		}
	}

	// 2. On prévient le Front que c'est fini
	msg := map[string]interface{}{
		"type": "GAME_OVER",
		"payload": map[string]interface{}{
			"message": "La partie est terminée !",
		},
	}
	data, _ := json.Marshal(msg)
	r.Broadcast <- data

	// 3. On remet les scores de tout le monde à 0
	r.resetScores()

	// 4. On renvoie les états mis à jour au Front (Lobby + Scores à 0)
	r.broadcastGameState()
	go r.BroadcastPlayerList()
}

// XPToLevel convertit un total d'XP en niveau selon la formule floor(sqrt(xp/100)) + 1.
func XPToLevel(xp int) int {
	if xp <= 0 {
		return 1
	}
	level := int(math.Sqrt(float64(xp)/100)) + 1
	if level < 1 {
		return 1
	}
	return level
}

// XPForScore calcule l'XP gagné pour un score de partie (minimum 5 pour la participation).
func XPForScore(score int) int {
	xp := score * 10
	if xp < 5 {
		return 5
	}
	return xp
}

// grantXP attribue de l'XP à chaque joueur connecté selon son score de la partie.
// XP gagné = score * 10 (minimum 5 pour la participation).
// Envoie un message XP_GAINED personnel à chaque joueur authentifié.
func (r *Room) grantXP() {
	for c := range r.Clients {
		if c.UserID == 0 || c.IsSpectator {
			continue
		}
		xpGained := XPForScore(c.Score)
		newXP, newLevel, err := database.AddUserXP(c.UserID, xpGained)
		if err != nil {
			log.Printf("Erreur AddUserXP pour %s: %v", c.Username, err)
			continue
		}
		if err := database.SaveGameResult(c.UserID, c.Score, xpGained); err != nil {
			log.Printf("Erreur SaveGameResult pour %s: %v", c.Username, err)
		}
		msg, _ := json.Marshal(map[string]interface{}{
			"type": "XP_GAINED",
			"payload": map[string]interface{}{
				"xp_gained": xpGained,
				"new_xp":    newXP,
				"new_level": newLevel,
			},
		})
		select {
		case c.Send <- msg:
		default:
		}
	}
}

func (r *Room) resetScores() {
	for c := range r.Clients {
		if !c.IsSpectator {
			c.Score = 0
		}
	}
}
