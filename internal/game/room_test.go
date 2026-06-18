package game

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/LeYapson/aniquiz/internal/models"
)

// newTestRoom crée une room avec des channels bufférisés pour éviter les deadlocks en test.
func newTestRoom() *Room {
	return &Room{
		ID:          "test-room",
		Clients:     make(map[*Client]bool),
		Broadcast:   make(chan []byte, 10),
		Register:    make(chan *Client, 1),
		Unregister:  make(chan *Client, 1),
		Start:       make(chan bool, 1),
		State:       StateLobby,
		HasAnswered: make(map[string]bool),
	}
}

func newTestClient(id, username string) *Client {
	return &Client{
		ID:       id,
		Username: username,
		Send:     make(chan []byte, 10),
		Score:    0,
	}
}

// --- CreateRoom ---

func TestCreateRoom_InitialState(t *testing.T) {
	room := CreateRoom("lobby-1", "creator-id", false)

	if room.ID != "lobby-1" {
		t.Errorf("ID: got %s, want lobby-1", room.ID)
	}
	if room.State != StateLobby {
		t.Errorf("State: got %s, want %s", room.State, StateLobby)
	}
	if room.IsPlaying {
		t.Error("IsPlaying devrait être false à la création")
	}
	if room.Clients == nil {
		t.Error("Clients map devrait être initialisée")
	}
}

// --- CheckAnswer ---

func TestCheckAnswer_GameNotStarted(t *testing.T) {
	room := newTestRoom()
	client := newTestClient("c1", "Alice")
	room.Clients[client] = true
	room.IsPlaying = false
	room.CurrentTrack = &models.Track{AnimeName: "Naruto Shippuden"}

	room.CheckAnswer(client, "Naruto Shippuden")

	if client.Score != 0 {
		t.Errorf("score: got %d, want 0 — partie non démarrée", client.Score)
	}
	select {
	case <-room.Broadcast:
		t.Error("aucun broadcast attendu quand la partie n'est pas démarrée")
	default:
	}
}

func TestCheckAnswer_NoTrack(t *testing.T) {
	room := newTestRoom()
	client := newTestClient("c1", "Alice")
	room.Clients[client] = true
	room.IsPlaying = true
	room.CurrentTrack = nil

	room.CheckAnswer(client, "Naruto Shippuden")

	if client.Score != 0 {
		t.Errorf("score: got %d, want 0 — pas de track", client.Score)
	}
}

func TestCheckAnswer_WrongAnswer(t *testing.T) {
	room := newTestRoom()
	client := newTestClient("c1", "Alice")
	room.Clients[client] = true
	room.IsPlaying = true
	room.CurrentTrack = &models.Track{AnimeName: "Naruto Shippuden"}

	room.CheckAnswer(client, "One Piece")

	if client.Score != 0 {
		t.Errorf("score: got %d, want 0 pour mauvaise réponse", client.Score)
	}
	select {
	case <-room.Broadcast:
		t.Error("aucun broadcast attendu pour une mauvaise réponse")
	default:
	}
}

func TestCheckAnswer_PartialAnswer(t *testing.T) {
	room := newTestRoom()
	client := newTestClient("c1", "Alice")
	room.Clients[client] = true
	room.IsPlaying = true
	room.CurrentTrack = &models.Track{AnimeName: "Naruto Shippuden"}

	room.CheckAnswer(client, "Naruto")

	// 5 pts (partielle) + 10 bonus premier = 15
	if client.Score != 15 {
		t.Errorf("score: got %d, want 15 pour réponse partielle (avec bonus premier)", client.Score)
	}
}

func TestCheckAnswer_CorrectAnswer_UpdatesScoreAndBroadcasts(t *testing.T) {
	room := newTestRoom()
	client := newTestClient("c1", "Alice")
	room.Clients[client] = true
	room.IsPlaying = true
	room.CurrentTrack = &models.Track{AnimeName: "Naruto Shippuden"}

	room.CheckAnswer(client, "Naruto Shippuden")

	// 10 pts (exacte) + 10 bonus premier = 20
	if client.Score != 20 {
		t.Errorf("score: got %d, want 20 pour bonne réponse (avec bonus premier)", client.Score)
	}

	select {
	case data := <-room.Broadcast:
		var msg models.WSMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			t.Fatalf("JSON invalide : %v", err)
		}
		if msg.Type != "PLAYER_GUESS" {
			t.Errorf("type du message: got %s, want PLAYER_GUESS", msg.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout : aucun message PLAYER_GUESS reçu")
	}
}

func TestCheckAnswer_ScoreAccumulates(t *testing.T) {
	room := newTestRoom()
	client := newTestClient("c1", "Alice")
	room.Clients[client] = true
	room.CurrentTrack = &models.Track{AnimeName: "Naruto Shippuden"}

	room.IsPlaying = true
	room.CheckAnswer(client, "Naruto") // +5 + 10 bonus premier = 15
	<-room.Broadcast                   // PLAYER_GUESS
	<-room.Broadcast                   // PLAYER_LIST de BroadcastPlayerList — garantit que la goroutine a fini de lire client.Score

	room.HasAnswered = make(map[string]bool) // simulate new round
	room.RoundAnswers = []RoundAnswer{}      // reset pour que le bonus premier s'applique à nouveau
	room.IsPlaying = true
	room.CheckAnswer(client, "Naruto Shippuden") // +10 + 10 bonus premier = 20
	<-room.Broadcast                             // PLAYER_GUESS
	<-room.Broadcast                             // PLAYER_LIST

	// Round 1 : 15 (5 partielle + 10 bonus) + Round 2 : 20 (10 exacte + 10 bonus) = 35
	if client.Score != 35 {
		t.Errorf("score cumulé: got %d, want 35", client.Score)
	}
}

// --- BroadcastPlayerList ---

func TestBroadcastPlayerList_MessageType(t *testing.T) {
	room := newTestRoom()
	client := newTestClient("c1", "Alice")
	client.Score = 42
	room.Clients[client] = true

	go room.BroadcastPlayerList()

	select {
	case data := <-room.Broadcast:
		var msg models.WSMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			t.Fatalf("JSON invalide : %v", err)
		}
		if msg.Type != "PLAYER_LIST" {
			t.Errorf("type du message: got %s, want PLAYER_LIST", msg.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout : BroadcastPlayerList n'a rien envoyé")
	}
}

func TestBroadcastPlayerList_EmptyRoom(t *testing.T) {
	room := newTestRoom()

	go room.BroadcastPlayerList()

	select {
	case data := <-room.Broadcast:
		var msg models.WSMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			t.Fatalf("JSON invalide : %v", err)
		}
		if msg.Type != "PLAYER_LIST" {
			t.Errorf("type du message: got %s, want PLAYER_LIST", msg.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout : BroadcastPlayerList n'a rien envoyé pour une room vide")
	}
}
