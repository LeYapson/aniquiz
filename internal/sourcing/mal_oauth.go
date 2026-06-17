package sourcing

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	malAuthURL  = "https://myanimelist.net/v1/oauth2/authorize"
	malTokenURL = "https://myanimelist.net/v1/oauth2/token"
	malAPIBase  = "https://api.myanimelist.net/v2"
)

func malClientID() string     { return os.Getenv("MAL_CLIENT_ID") }
func malClientSecret() string { return os.Getenv("MAL_CLIENT_SECRET") }
func malRedirectURI() string {
	if u := os.Getenv("MAL_REDIRECT_URI"); u != "" {
		return u
	}
	return "http://localhost:8080/api/auth/mal/callback"
}

// malState est encodé en base64 dans le paramètre state pour rester stateless.
type malState struct {
	JWT      string `json:"j"`
	Verifier string `json:"v"`
}

func encodeMalState(jwt, verifier string) string {
	data, _ := json.Marshal(malState{JWT: jwt, Verifier: verifier})
	return base64.RawURLEncoding.EncodeToString(data)
}

func DecodeMalState(state string) (jwt, verifier string, err error) {
	data, err := base64.RawURLEncoding.DecodeString(state)
	if err != nil {
		return "", "", fmt.Errorf("state invalide : %w", err)
	}
	var s malState
	if err := json.Unmarshal(data, &s); err != nil {
		return "", "", fmt.Errorf("state mal formé : %w", err)
	}
	return s.JWT, s.Verifier, nil
}

// generateCodeVerifier génère un code_verifier PKCE aléatoire (RFC 7636).
func generateCodeVerifier() (string, error) {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// BuildMALAuthURL construit l'URL OAuth MAL avec PKCE (méthode plain, seule supportée par MAL).
func BuildMALAuthURL(jwtToken string) (authURL, verifier string, err error) {
	verifier, err = generateCodeVerifier()
	if err != nil {
		return "", "", fmt.Errorf("impossible de générer le verifier PKCE : %w", err)
	}
	state := encodeMalState(jwtToken, verifier)

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", malClientID())
	params.Set("redirect_uri", malRedirectURI())
	params.Set("state", state)
	params.Set("code_challenge", verifier)
	params.Set("code_challenge_method", "plain")

	return malAuthURL + "?" + params.Encode(), verifier, nil
}

type malTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// ExchangeMALCode échange le code d'autorisation MAL contre un access token via PKCE.
func ExchangeMALCode(code, verifier string) (string, error) {
	data := url.Values{}
	data.Set("client_id", malClientID())
	data.Set("client_secret", malClientSecret())
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", malRedirectURI())
	data.Set("code_verifier", verifier)

	resp, err := http.Post(malTokenURL, "application/x-www-form-urlencoded", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("échange de code MAL échoué : %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("MAL a refusé le code (%d) : %s", resp.StatusCode, string(raw))
	}

	var result malTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("impossible de décoder la réponse MAL : %w", err)
	}
	return result.AccessToken, nil
}

// GetMALAnimeList retourne les MAL IDs de la liste de l'utilisateur (completed + watching).
func GetMALAnimeList(token string) ([]int, error) {
	seen := make(map[int]bool)
	var ids []int

	for _, status := range []string{"completed", "watching"} {
		apiURL := malAPIBase + "/users/@me/animelist?fields=list_status&status=" + status + "&limit=1000"
		req, _ := http.NewRequest(http.MethodGet, apiURL, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		var result struct {
			Data []struct {
				Node struct {
					ID int `json:"id"`
				} `json:"node"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			continue
		}
		for _, item := range result.Data {
			if !seen[item.Node.ID] {
				seen[item.Node.ID] = true
				ids = append(ids, item.Node.ID)
			}
		}
	}
	return ids, nil
}

type MALProfile struct {
	ID       int
	Username string
}

// GetMALProfile récupère le profil de l'utilisateur via son access token MAL.
func GetMALProfile(token string) (*MALProfile, error) {
	req, _ := http.NewRequest(http.MethodGet, malAPIBase+"/users/@me?fields=id,name", strings.NewReader(""))
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("appel API MAL échoué : %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("impossible de décoder le profil MAL : %w", err)
	}
	return &MALProfile{ID: result.ID, Username: result.Name}, nil
}
