package sourcing

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ThemesResponse []struct {
	Themes []struct {
		ThemeType string `json:"themeType"`
		Mirror    struct {
			MirrorURL string `json:"mirrorURL"`
		} `json:"mirror"`
	} `json:"themes"`
}

func GetAudioURL(malID int) (map[string]string, error) {
	url := fmt.Sprintf("https://themes.moe/api/themes/%d", malID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 1. On lit les octets UNE SEULE FOIS
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 2. On utilise json.Unmarshal sur les octets qu'on a en mémoire (body)
	var apiData ThemesResponse
	if err := json.Unmarshal(body, &apiData); err != nil {
		return nil, fmt.Errorf("erreur décodage: %v", err)
	}

	links := make(map[string]string)
	if len(apiData) > 0 {
		for _, t := range apiData[0].Themes {
			if t.Mirror.MirrorURL != "" {
				links[t.ThemeType] = t.Mirror.MirrorURL
				fmt.Printf("✅ Trouvé [%s] : %s\n", t.ThemeType, t.Mirror.MirrorURL)
			}
		}
	}
	return links, nil
}