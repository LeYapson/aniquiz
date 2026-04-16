package sourcing

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Structure pour décoder la réponse de themes.moe
type ThemesResponse []struct {
	Themes []struct {
		Type string `json:"type"`
		Mirror struct {
			MirrorURL string `json:"mirror_url"`
		} `json:"mirror"`
	} `json:"themes"`
}

func GetAudioURL(malID int) (map[string]string, error) {
	//l'URL de l'API (exemple basé sur le fonctionnement classique de themes.moe)
	url := fmt.Sprintf("https://themes.moe/api/themes/%d", malID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiData ThemesResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiData); err != nil {
		return nil, err
	}

	links := make(map[string]string)
	if len(apiData) > 0 {
		for _, t := range apiData[0].Themes {
			//on stocke les liens dans une map avec le type (opening/ending) comme clé
			links[t.Type] = t.Mirror.MirrorURL
		}
	}
	return links, nil
}