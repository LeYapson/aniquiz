package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LeYapson/aniquiz/internal/game"
	"github.com/LeYapson/aniquiz/internal/handlers"
	"github.com/LeYapson/aniquiz/internal/models"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// mockStore implémente handlers.Store sans base de données réelle.
type mockStore struct {
	track *models.Track
	err   error
}

func (m *mockStore) GetRandomTrack() (*models.Track, error)    { return m.track, m.err }
func (m *mockStore) GetTrackByID(_ int) (*models.Track, error) { return m.track, m.err }
func (m *mockStore) GetAllTracks() ([]models.Track, error) {
	if m.track != nil {
		return []models.Track{*m.track}, m.err
	}
	return nil, m.err
}

// helpers

func newServer(store handlers.Store) *gin.Engine {
	return handlers.NewRouter(store)
}

func doJSON(router *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// --- GET /ping ---

func TestPing(t *testing.T) {
	router := newServer(&mockStore{})
	w := doJSON(router, http.MethodGet, "/ping", nil)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["message"] != "pong" {
		t.Errorf("message: got %v, want pong", resp["message"])
	}
}

// --- POST /rooms ---

func cleanRoom(id string) {
	game.RoomsMu.Lock()
	delete(game.ActiveRooms, id)
	game.RoomsMu.Unlock()
}

func TestCreateRoom_OK(t *testing.T) {
	defer cleanRoom("test-salon")
	router := newServer(&mockStore{})

	w := doJSON(router, http.MethodPost, "/rooms", map[string]string{"room_id": "test-salon"})

	if w.Code != http.StatusCreated {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusCreated)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["room_id"] != "test-salon" {
		t.Errorf("room_id: got %v, want test-salon", resp["room_id"])
	}
}

func TestCreateRoom_MissingRoomID(t *testing.T) {
	router := newServer(&mockStore{})

	w := doJSON(router, http.MethodPost, "/rooms", map[string]string{})

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateRoom_Duplicate(t *testing.T) {
	defer cleanRoom("salon-doublon")
	router := newServer(&mockStore{})

	doJSON(router, http.MethodPost, "/rooms", map[string]string{"room_id": "salon-doublon"})
	w := doJSON(router, http.MethodPost, "/rooms", map[string]string{"room_id": "salon-doublon"})

	if w.Code != http.StatusConflict {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusConflict)
	}
}

// --- GET /quiz/next ---

func TestQuizNext_OK(t *testing.T) {
	track := &models.Track{ID: 7, AudioURL: "https://example.com/track.webm", AnimeName: "Naruto"}
	router := newServer(&mockStore{track: track})

	w := doJSON(router, http.MethodGet, "/quiz/next", nil)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	question, ok := resp["question"].(map[string]any)
	if !ok {
		t.Fatal("champ 'question' absent ou mal formé")
	}
	if question["audio_url"] != track.AudioURL {
		t.Errorf("audio_url: got %v, want %s", question["audio_url"], track.AudioURL)
	}
	// Le nom de l'anime ne doit PAS être exposé dans la question
	if _, exposed := question["anime_name"]; exposed {
		t.Error("anime_name ne doit pas être exposé dans la question")
	}
}

func TestQuizNext_DBError(t *testing.T) {
	router := newServer(&mockStore{err: errors.New("connexion DB perdue")})

	w := doJSON(router, http.MethodGet, "/quiz/next", nil)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

// --- POST /quiz/answer ---

func TestQuizAnswer_InvalidJSON(t *testing.T) {
	router := newServer(&mockStore{})
	req := httptest.NewRequest(http.MethodPost, "/quiz/answer", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestQuizAnswer_TrackNotFound(t *testing.T) {
	router := newServer(&mockStore{err: errors.New("not found")})

	w := doJSON(router, http.MethodPost, "/quiz/answer", map[string]any{
		"track_id": 999,
		"answer":   "Naruto",
	})

	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestQuizAnswer_CorrectAnswer(t *testing.T) {
	track := &models.Track{ID: 1, AnimeName: "Naruto Shippuden"}
	router := newServer(&mockStore{track: track})

	w := doJSON(router, http.MethodPost, "/quiz/answer", map[string]any{
		"track_id": 1,
		"answer":   "Naruto Shippuden",
	})

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["correct"] != true {
		t.Errorf("correct: got %v, want true", resp["correct"])
	}
	if resp["points"] != float64(10) {
		t.Errorf("points: got %v, want 10", resp["points"])
	}
}

func TestQuizAnswer_WrongAnswer(t *testing.T) {
	track := &models.Track{ID: 1, AnimeName: "Naruto Shippuden"}
	router := newServer(&mockStore{track: track})

	w := doJSON(router, http.MethodPost, "/quiz/answer", map[string]any{
		"track_id": 1,
		"answer":   "One Piece",
	})

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["correct"] != false {
		t.Errorf("correct: got %v, want false", resp["correct"])
	}
	if resp["points"] != float64(0) {
		t.Errorf("points: got %v, want 0", resp["points"])
	}
}

func TestQuizAnswer_PartialAnswer(t *testing.T) {
	track := &models.Track{ID: 1, AnimeName: "Naruto Shippuden"}
	router := newServer(&mockStore{track: track})

	w := doJSON(router, http.MethodPost, "/quiz/answer", map[string]any{
		"track_id": 1,
		"answer":   "Naruto",
	})

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["points"] != float64(5) {
		t.Errorf("points: got %v, want 5 pour réponse partielle", resp["points"])
	}
}
