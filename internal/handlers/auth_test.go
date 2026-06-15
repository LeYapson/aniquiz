package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LeYapson/aniquiz/internal/handlers"
	"github.com/LeYapson/aniquiz/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// authMockStore offre un contrôle fin sur les réponses d'auth.
type authMockStore struct {
	existingUser  *models.User
	createErr     error
	getUserErr    error
}

func (m *authMockStore) GetRandomTrack() (*models.Track, error)    { return nil, nil }
func (m *authMockStore) GetTrackByID(_ int) (*models.Track, error) { return nil, nil }
func (m *authMockStore) GetAllTracks() ([]models.Track, error)     { return nil, nil }
func (m *authMockStore) CreateUser(_, _, _ string) error           { return m.createErr }
func (m *authMockStore) GetUserByUsernameOrEmail(_ string) (*models.User, error) {
	if m.getUserErr != nil {
		return nil, m.getUserErr
	}
	return m.existingUser, nil
}

// --- RegisterHandler ---

func TestRegisterHandler_Success(t *testing.T) {
	store := &authMockStore{getUserErr: errors.New("not found")}
	router := handlers.NewRouter(store)

	w := doJSON(router, http.MethodPost, "/api/auth/register", map[string]string{
		"username": "newuser",
		"email":    "new@example.com",
		"password": "password123",
	})

	if w.Code != http.StatusCreated {
		t.Errorf("status: got %d, want %d — body: %s", w.Code, http.StatusCreated, w.Body.String())
	}
}

func TestRegisterHandler_InvalidData(t *testing.T) {
	router := handlers.NewRouter(&authMockStore{})

	// Mot de passe trop court
	w := doJSON(router, http.MethodPost, "/api/auth/register", map[string]string{
		"username": "newuser",
		"email":    "new@example.com",
		"password": "short",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestRegisterHandler_DuplicateUsername(t *testing.T) {
	existing := &models.User{ID: 1, Username: "existing", Email: "existing@example.com"}
	store := &authMockStore{existingUser: existing}
	router := handlers.NewRouter(store)

	w := doJSON(router, http.MethodPost, "/api/auth/register", map[string]string{
		"username": "existing",
		"email":    "other@example.com",
		"password": "password123",
	})
	if w.Code != http.StatusConflict {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusConflict)
	}
}

// --- LoginHandler ---

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("impossible de hasher le mot de passe: %v", err)
	}
	return string(h)
}

func TestLoginHandler_Success(t *testing.T) {
	user := &models.User{
		ID:           1,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashPassword(t, "password123"),
	}
	store := &authMockStore{existingUser: user}
	router := handlers.NewRouter(store)

	w := doJSON(router, http.MethodPost, "/api/auth/login", map[string]string{
		"identifier": "testuser",
		"password":   "password123",
	})

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d — body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["token"] == nil || resp["token"] == "" {
		t.Error("le token doit être présent dans la réponse de login")
	}
	if resp["user"] == nil {
		t.Error("les infos utilisateur doivent être présentes")
	}
}

func TestLoginHandler_WrongPassword(t *testing.T) {
	user := &models.User{
		ID:           1,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashPassword(t, "password123"),
	}
	store := &authMockStore{existingUser: user}
	router := handlers.NewRouter(store)

	w := doJSON(router, http.MethodPost, "/api/auth/login", map[string]string{
		"identifier": "testuser",
		"password":   "wrongpassword",
	})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestLoginHandler_UnknownUser(t *testing.T) {
	store := &authMockStore{getUserErr: errors.New("not found")}
	router := handlers.NewRouter(store)

	w := doJSON(router, http.MethodPost, "/api/auth/login", map[string]string{
		"identifier": "nobody",
		"password":   "password123",
	})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// --- AuthMiddleware ---

func newProtectedRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	protected := r.Group("/")
	protected.Use(handlers.AuthMiddleware())
	protected.GET("/protected", func(c *gin.Context) {
		username, _ := c.Get("username")
		c.JSON(http.StatusOK, gin.H{"username": username})
	})
	return r
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	router := newProtectedRouter()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	router := newProtectedRouter()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	router := newProtectedRouter()
	tok, _ := handlers.GenerateToken(1, "testuser")

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["username"] != "testuser" {
		t.Errorf("username: got %v, want testuser", resp["username"])
	}
}
