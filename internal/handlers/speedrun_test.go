package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/LeYapson/aniquiz/internal/handlers"
	"github.com/LeYapson/aniquiz/internal/models"
)

// speedrunMockStore offre un contrôle fin sur les réponses spécifiques au speed run.
type speedrunMockStore struct {
	track          *models.Track
	trackErr       error
	saveResultErr  error
	leaderboard    []models.SpeedrunLeaderboardEntry
	leaderboardErr error
	// callCount permet de vérifier que SaveSpeedrunResult a bien été appelé.
	saveResultCalls int
}

func (m *speedrunMockStore) GetRandomTrack() (*models.Track, error)    { return m.track, m.trackErr }
func (m *speedrunMockStore) GetTrackByID(_ int) (*models.Track, error) { return nil, nil }
func (m *speedrunMockStore) GetAllTracks() ([]models.Track, error)     { return nil, nil }
func (m *speedrunMockStore) CreateUser(_, _, _ string) error           { return nil }
func (m *speedrunMockStore) GetUserByUsernameOrEmail(_ string) (*models.User, error) {
	return nil, nil
}
func (m *speedrunMockStore) SaveSpeedrunResult(_, _ int) error {
	m.saveResultCalls++
	return m.saveResultErr
}
func (m *speedrunMockStore) GetSpeedrunLeaderboard(_ int) ([]models.SpeedrunLeaderboardEntry, error) {
	return m.leaderboard, m.leaderboardErr
}

// startSession est un helper qui appelle POST /api/speedrun/start et retourne le session_id.
func startSession(t *testing.T, store handlers.Store) string {
	t.Helper()
	router := newServer(store)
	tok := testToken(t)
	w := doJSONAuth(router, http.MethodPost, "/api/speedrun/start", nil, tok)
	if w.Code != http.StatusOK {
		t.Fatalf("startSession: status %d — %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	id, ok := resp["session_id"].(string)
	if !ok || id == "" {
		t.Fatal("startSession: session_id absent")
	}
	return id
}

// --- POST /api/speedrun/start ---

func TestSpeedrunStart_OK(t *testing.T) {
	track := &models.Track{ID: 1, AudioURL: "https://example.com/op.webm", AnimeName: "Naruto"}
	store := &speedrunMockStore{track: track}
	router := newServer(store)
	tok := testToken(t)

	w := doJSONAuth(router, http.MethodPost, "/api/speedrun/start", nil, tok)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d — %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["session_id"] == nil || resp["session_id"] == "" {
		t.Error("session_id doit être présent")
	}
	if resp["expires_at"] == nil {
		t.Error("expires_at doit être présent")
	}
	trackObj, ok := resp["track"].(map[string]any)
	if !ok {
		t.Fatal("track doit être présent")
	}
	if trackObj["audio_url"] != track.AudioURL {
		t.Errorf("audio_url: got %v, want %s", trackObj["audio_url"], track.AudioURL)
	}
	// La réponse correcte ne doit pas être exposée
	if trackObj["anime_name"] != nil {
		t.Error("anime_name ne doit pas être exposé dans la réponse start")
	}
}

func TestSpeedrunStart_Unauthenticated(t *testing.T) {
	store := &speedrunMockStore{track: &models.Track{ID: 1}}
	router := newServer(store)

	w := doJSON(router, http.MethodPost, "/api/speedrun/start", nil)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestSpeedrunStart_DBError(t *testing.T) {
	store := &speedrunMockStore{trackErr: errors.New("connexion DB perdue")}
	router := newServer(store)
	tok := testToken(t)

	w := doJSONAuth(router, http.MethodPost, "/api/speedrun/start", nil, tok)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

// --- POST /api/speedrun/answer ---

func TestSpeedrunAnswer_MissingBody(t *testing.T) {
	track := &models.Track{ID: 1, AnimeName: "Naruto", AudioURL: "x"}
	store := &speedrunMockStore{track: track}
	router := newServer(store)
	tok := testToken(t)

	w := doJSONAuth(router, http.MethodPost, "/api/speedrun/answer", map[string]string{
		"answer": "Naruto",
		// session_id intentionnellement absent
	}, tok)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestSpeedrunAnswer_InvalidSession(t *testing.T) {
	store := &speedrunMockStore{track: &models.Track{ID: 1}}
	router := newServer(store)
	tok := testToken(t)

	w := doJSONAuth(router, http.MethodPost, "/api/speedrun/answer", map[string]string{
		"session_id": "session-qui-nexiste-pas",
		"answer":     "Naruto",
	}, tok)

	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestSpeedrunAnswer_WrongAnswer(t *testing.T) {
	track := &models.Track{ID: 1, AnimeName: "Naruto Shippuden", AudioURL: "x"}
	store := &speedrunMockStore{track: track}
	router := newServer(store)
	tok := testToken(t)

	sessionID := startSession(t, store)

	w := doJSONAuth(router, http.MethodPost, "/api/speedrun/answer", map[string]string{
		"session_id": sessionID,
		"answer":     "One Piece",
	}, tok)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["correct"] != false {
		t.Errorf("correct: got %v, want false", resp["correct"])
	}
	if resp["next_track"] != nil {
		t.Error("next_track ne doit pas être présent sur mauvaise réponse")
	}
}

func TestSpeedrunAnswer_CorrectAnswer(t *testing.T) {
	track := &models.Track{ID: 1, AnimeName: "Naruto Shippuden", AudioURL: "x"}
	nextTrack := &models.Track{ID: 2, AnimeName: "One Piece", AudioURL: "y"}
	callCount := 0
	store := &speedrunMockStore{}
	store.track = track
	// Premier appel (start) → track, deuxième appel (next après bonne réponse) → nextTrack
	origGetRandom := store.track
	_ = origGetRandom
	// On utilise un store plus flexible pour simuler deux tracks différents
	flexStore := &flexSpeedrunStore{
		tracks: []*models.Track{track, nextTrack},
		idx:    &callCount,
	}

	router := newServer(flexStore)
	tok := testToken(t)
	sessionID := startSession(t, flexStore)

	w := doJSONAuth(router, http.MethodPost, "/api/speedrun/answer", map[string]string{
		"session_id": sessionID,
		"answer":     "Naruto Shippuden",
	}, tok)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d — %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["correct"] != true {
		t.Errorf("correct: got %v, want true", resp["correct"])
	}
	if resp["score"] != float64(1) {
		t.Errorf("score: got %v, want 1", resp["score"])
	}
	nextObj, ok := resp["next_track"].(map[string]any)
	if !ok {
		t.Fatal("next_track doit être présent sur bonne réponse")
	}
	if nextObj["audio_url"] != nextTrack.AudioURL {
		t.Errorf("next_track.audio_url: got %v, want %s", nextObj["audio_url"], nextTrack.AudioURL)
	}
}

// flexSpeedrunStore alterne entre plusieurs tracks pour simuler des appels successifs.
type flexSpeedrunStore struct {
	tracks          []*models.Track
	idx             *int
	saveResultCalls int
	leaderboard     []models.SpeedrunLeaderboardEntry
}

func (f *flexSpeedrunStore) GetRandomTrack() (*models.Track, error) {
	i := *f.idx % len(f.tracks)
	*f.idx++
	return f.tracks[i], nil
}
func (f *flexSpeedrunStore) GetTrackByID(_ int) (*models.Track, error) { return nil, nil }
func (f *flexSpeedrunStore) GetAllTracks() ([]models.Track, error)     { return nil, nil }
func (f *flexSpeedrunStore) CreateUser(_, _, _ string) error           { return nil }
func (f *flexSpeedrunStore) GetUserByUsernameOrEmail(_ string) (*models.User, error) {
	return nil, nil
}
func (f *flexSpeedrunStore) SaveSpeedrunResult(_, _ int) error {
	f.saveResultCalls++
	return nil
}
func (f *flexSpeedrunStore) GetSpeedrunLeaderboard(_ int) ([]models.SpeedrunLeaderboardEntry, error) {
	return f.leaderboard, nil
}

// --- POST /api/speedrun/skip ---

func TestSpeedrunSkip_InvalidSession(t *testing.T) {
	store := &speedrunMockStore{track: &models.Track{ID: 1, AudioURL: "x"}}
	router := newServer(store)
	tok := testToken(t)

	w := doJSONAuth(router, http.MethodPost, "/api/speedrun/skip", map[string]string{
		"session_id": "inexistant",
	}, tok)

	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestSpeedrunSkip_OK(t *testing.T) {
	callCount := 0
	track1 := &models.Track{ID: 1, AnimeName: "Naruto", AudioURL: "url1"}
	track2 := &models.Track{ID: 2, AnimeName: "One Piece", AudioURL: "url2"}
	store := &flexSpeedrunStore{tracks: []*models.Track{track1, track2}, idx: &callCount}
	router := newServer(store)
	tok := testToken(t)

	sessionID := startSession(t, store)

	w := doJSONAuth(router, http.MethodPost, "/api/speedrun/skip", map[string]string{
		"session_id": sessionID,
	}, tok)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d — %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["skipped"] != true {
		t.Error("skipped doit être true")
	}
	if resp["score"] != float64(0) {
		t.Errorf("score: got %v, want 0 (skip ne rapporte pas de points)", resp["score"])
	}
	nextObj, ok := resp["next_track"].(map[string]any)
	if !ok {
		t.Fatal("next_track doit être présent")
	}
	if nextObj["audio_url"] != track2.AudioURL {
		t.Errorf("next_track.audio_url: got %v, want %s", nextObj["audio_url"], track2.AudioURL)
	}
}

func TestSpeedrunSkip_MissingSessionID(t *testing.T) {
	store := &speedrunMockStore{track: &models.Track{ID: 1, AudioURL: "x"}}
	router := newServer(store)
	tok := testToken(t)

	w := doJSONAuth(router, http.MethodPost, "/api/speedrun/skip", map[string]string{}, tok)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// --- POST /api/speedrun/finish ---

func TestSpeedrunFinish_InvalidSession(t *testing.T) {
	store := &speedrunMockStore{track: &models.Track{ID: 1, AudioURL: "x"}}
	router := newServer(store)
	tok := testToken(t)

	w := doJSONAuth(router, http.MethodPost, "/api/speedrun/finish", map[string]string{
		"session_id": "inexistant",
	}, tok)

	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestSpeedrunFinish_OK(t *testing.T) {
	track := &models.Track{ID: 1, AnimeName: "Naruto", AudioURL: "x"}
	store := &speedrunMockStore{track: track}
	router := newServer(store)
	tok := testToken(t)

	sessionID := startSession(t, store)

	w := doJSONAuth(router, http.MethodPost, "/api/speedrun/finish", map[string]string{
		"session_id": sessionID,
	}, tok)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d — %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["score"] == nil {
		t.Error("score doit être présent dans la réponse finish")
	}
	if store.saveResultCalls != 1 {
		t.Errorf("SaveSpeedrunResult appelé %d fois, attendu 1", store.saveResultCalls)
	}
}

func TestSpeedrunFinish_DoubleCalls(t *testing.T) {
	track := &models.Track{ID: 1, AnimeName: "Naruto", AudioURL: "x"}
	store := &speedrunMockStore{track: track}
	router := newServer(store)
	tok := testToken(t)

	sessionID := startSession(t, store)

	// Premier appel — OK
	doJSONAuth(router, http.MethodPost, "/api/speedrun/finish", map[string]string{
		"session_id": sessionID,
	}, tok)

	// Deuxième appel sur la même session — doit retourner 404
	w := doJSONAuth(router, http.MethodPost, "/api/speedrun/finish", map[string]string{
		"session_id": sessionID,
	}, tok)

	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d (session déjà supprimée)", w.Code, http.StatusNotFound)
	}
}

func TestSpeedrunFinish_SaveError(t *testing.T) {
	track := &models.Track{ID: 1, AnimeName: "Naruto", AudioURL: "x"}
	store := &speedrunMockStore{track: track, saveResultErr: errors.New("DB error")}
	router := newServer(store)
	tok := testToken(t)

	sessionID := startSession(t, store)

	w := doJSONAuth(router, http.MethodPost, "/api/speedrun/finish", map[string]string{
		"session_id": sessionID,
	}, tok)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

// --- GET /api/leaderboard/speedrun ---

func TestSpeedrunLeaderboard_OK(t *testing.T) {
	entries := []models.SpeedrunLeaderboardEntry{
		{Rank: 1, UserID: 1, Username: "alice", BestScore: 42},
		{Rank: 2, UserID: 2, Username: "bob", BestScore: 35},
	}
	store := &speedrunMockStore{
		track:       &models.Track{ID: 1},
		leaderboard: entries,
	}
	router := newServer(store)

	w := doJSON(router, http.MethodGet, "/api/leaderboard/speedrun", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d — %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp []map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	if len(resp) != 2 {
		t.Fatalf("nombre d'entrées: got %d, want 2", len(resp))
	}
	if resp[0]["username"] != "alice" {
		t.Errorf("premier joueur: got %v, want alice", resp[0]["username"])
	}
}

func TestSpeedrunLeaderboard_Empty(t *testing.T) {
	store := &speedrunMockStore{track: &models.Track{ID: 1}, leaderboard: nil}
	router := newServer(store)

	w := doJSON(router, http.MethodGet, "/api/leaderboard/speedrun", nil)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp []any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 0 {
		t.Errorf("attendu tableau vide, got %v", resp)
	}
}

func TestSpeedrunLeaderboard_DBError(t *testing.T) {
	store := &speedrunMockStore{
		track:          &models.Track{ID: 1},
		leaderboardErr: errors.New("DB error"),
	}
	router := newServer(store)

	w := doJSON(router, http.MethodGet, "/api/leaderboard/speedrun", nil)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
