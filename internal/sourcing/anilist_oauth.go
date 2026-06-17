package sourcing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	anilistAuthURL  = "https://anilist.co/api/v2/oauth/authorize"
	anilistTokenURL = "https://anilist.co/api/v2/oauth/token"
	anilistGraphQL  = "https://graphql.anilist.co"
)

func anilistClientID() string     { return os.Getenv("ANILIST_CLIENT_ID") }
func anilistClientSecret() string { return os.Getenv("ANILIST_CLIENT_SECRET") }
func anilistRedirectURI() string {
	if u := os.Getenv("ANILIST_REDIRECT_URI"); u != "" {
		return u
	}
	return "http://localhost:8080/api/auth/anilist/callback"
}

// BuildAuthURL construit l'URL de redirection vers AniList.
// state contient le JWT de l'utilisateur pour retrouver son compte au retour.
func BuildAuthURL(state string) string {
	return fmt.Sprintf(
		"%s?client_id=%s&redirect_uri=%s&response_type=code&state=%s",
		anilistAuthURL,
		anilistClientID(),
		anilistRedirectURI(),
		state,
	)
}

type anilistTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// ExchangeCode échange le code d'autorisation AniList contre un access token.
func ExchangeCode(code string) (string, error) {
	payload := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     anilistClientID(),
		"client_secret": anilistClientSecret(),
		"redirect_uri":  anilistRedirectURI(),
		"code":          code,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(anilistTokenURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("échange de code AniList échoué : %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("AniList a refusé le code (%d) : %s", resp.StatusCode, string(raw))
	}

	var result anilistTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("impossible de décoder la réponse AniList : %w", err)
	}
	return result.AccessToken, nil
}

// GetAnilistAnimeList retourne les MAL IDs de la liste COMPLETED + WATCHING de l'utilisateur.
func GetAnilistAnimeList(token string, anilistUserID int) ([]int, error) {
	query := `
	query ($userId: Int) {
	  MediaListCollection(userId: $userId, type: ANIME, status_in: [COMPLETED, CURRENT]) {
	    lists {
	      entries {
	        media { idMal }
	      }
	    }
	  }
	}`
	payload := map[string]interface{}{
		"query":     query,
		"variables": map[string]interface{}{"userId": anilistUserID},
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, anilistGraphQL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("appel GraphQL AniList anime list échoué : %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			MediaListCollection struct {
				Lists []struct {
					Entries []struct {
						Media struct {
							IDMal *int `json:"idMal"`
						} `json:"media"`
					} `json:"entries"`
				} `json:"lists"`
			} `json:"MediaListCollection"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("décodage liste AniList échoué : %w", err)
	}

	seen := make(map[int]bool)
	var ids []int
	for _, list := range result.Data.MediaListCollection.Lists {
		for _, entry := range list.Entries {
			if entry.Media.IDMal != nil && !seen[*entry.Media.IDMal] {
				seen[*entry.Media.IDMal] = true
				ids = append(ids, *entry.Media.IDMal)
			}
		}
	}
	return ids, nil
}

type AnilistProfile struct {
	ID       int
	Username string
}

// GetAnilistProfile récupère l'identifiant et le pseudo AniList de l'utilisateur via son token.
func GetAnilistProfile(token string) (*AnilistProfile, error) {
	query := `{ Viewer { id name } }`
	payload := map[string]string{"query": query}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, anilistGraphQL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("appel GraphQL AniList échoué : %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Viewer struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"Viewer"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("impossible de décoder le profil AniList : %w", err)
	}
	return &AnilistProfile{
		ID:       result.Data.Viewer.ID,
		Username: result.Data.Viewer.Name,
	}, nil
}
