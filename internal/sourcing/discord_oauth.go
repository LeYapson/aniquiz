package sourcing

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	discordAuthURL  = "https://discord.com/api/oauth2/authorize"
	discordTokenURL = "https://discord.com/api/oauth2/token"
	discordMeURL    = "https://discord.com/api/users/@me"
)

func discordClientID() string     { return os.Getenv("DISCORD_CLIENT_ID") }
func discordClientSecret() string { return os.Getenv("DISCORD_CLIENT_SECRET") }
func discordRedirectURI() string {
	if u := os.Getenv("DISCORD_REDIRECT_URI"); u != "" {
		return u
	}
	return "http://localhost:8080/api/auth/discord/callback"
}

// BuildDiscordAuthURL construit l'URL d'autorisation Discord (scope identify).
// state porte le JWT de l'utilisateur pour le retrouver au callback.
func BuildDiscordAuthURL(state string) string {
	q := url.Values{}
	q.Set("client_id", discordClientID())
	q.Set("redirect_uri", discordRedirectURI())
	q.Set("response_type", "code")
	q.Set("scope", "identify")
	q.Set("state", state)
	return discordAuthURL + "?" + q.Encode()
}

type discordTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// ExchangeDiscordCode échange le code d'autorisation contre un access token.
// Discord attend un corps form-urlencoded (≠ AniList qui accepte du JSON).
func ExchangeDiscordCode(code string) (string, error) {
	form := url.Values{}
	form.Set("client_id", discordClientID())
	form.Set("client_secret", discordClientSecret())
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", discordRedirectURI())

	resp, err := http.Post(discordTokenURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("échange de code Discord échoué : %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Discord a refusé le code (%d) : %s", resp.StatusCode, string(raw))
	}

	var result discordTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("impossible de décoder la réponse Discord : %w", err)
	}
	return result.AccessToken, nil
}

// DiscordProfile : identité Discord minimale dont on a besoin pour le lien.
type DiscordProfile struct {
	ID       string
	Username string
}

// GetDiscordProfile récupère l'id et le pseudo Discord via l'access token.
func GetDiscordProfile(token string) (*DiscordProfile, error) {
	req, _ := http.NewRequest(http.MethodGet, discordMeURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("appel profil Discord échoué : %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("profil Discord refusé (%d) : %s", resp.StatusCode, string(raw))
	}

	var u struct {
		ID         string `json:"id"`
		Username   string `json:"username"`
		GlobalName string `json:"global_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("décodage profil Discord échoué : %w", err)
	}

	name := u.GlobalName // pseudo d'affichage récent ; fallback sur username
	if name == "" {
		name = u.Username
	}
	return &DiscordProfile{ID: u.ID, Username: name}, nil
}
