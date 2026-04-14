package sourcing

import (
	"fmt"
	"strings"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/LeYapson/aniquiz/internal/models"
	"github.com/darenliang/jikan-go"
)

type AnimeMusicInfo struct {
	Title string
	Openings []string
	Endings []string

}

func ProcessAndSaveAnime(animeId int) (*AnimeMusicInfo, error) {
    // 1. Récupération Jikan
    musicInfo, err := GetAnimeMusic(animeId)
    if err != nil {
        return nil, err
    }

    // 2. Récupération Audio (Linker)
    audiolinks, _ := GetAudioURL(animeId)

    // 3. Boucle sur les Openings
    for i, op := range musicInfo.Openings {
        // NETTOYAGE : On utilise ta fonction parseTrack
        cleanTitle, cleanArtist := parseTrack(op)

        // AUDIO : Themes.moe utilise "OP1", "OP2"... pas "opening1"
        opKey := fmt.Sprintf("OP%d", i+1) 
        audioURL := "not_found"
        if url, ok := audiolinks[opKey]; ok {
            audioURL = url
        }

        track := models.Track{
            Title:      cleanTitle,
            Artist:     cleanArtist,
            AnimeName:  musicInfo.Title,
            AudioURL:   audioURL,
            MalID:      animeId,
            Difficulty: 1, // On force à 1 pour éviter le 0 par défaut
        }
        
        database.SaveTrack(track)
    }

    return musicInfo, nil
}

func GetAnimeMusic(animeId int) (*AnimeMusicInfo, error) {
	//on recupere les themes specifiques*
	themes, err := jikan.GetAnimeThemes(animeId)
	if err != nil {
		return nil, err
	}

	//On recupere les infos de base (pour le titre)
	anime, err := jikan.GetAnimeById(animeId)
	if err != nil {
		return nil, err
	}

	return &AnimeMusicInfo{
		Title: anime.Data.Title,
		Openings: themes.Data.Openings,
		Endings: themes.Data.Endings,
	}, nil
}

// parseTrack transforme '"Tank!" by The Seatbelts (eps 1-25)' 
// en Titre: Tank!, Artiste: The Seatbelts
func parseTrack(rawTitle string) (string, string) {
	// 1 - retire les episodes (ce qui est entre parenthèses)
	parts := strings.Split(rawTitle, "(")
	titleAndArtist := parts[0]

	//2 - on sépare le titre et l'artiste par le mot "by"
	parts = strings.Split(titleAndArtist, "by")

	title := strings.Trim(parts[0], " \"") // on retire les espaces et les guillemets
	artist := "Unknown Artist"
	if len(parts) > 1 {
		artist = parts[1]
	}

	return title, artist
}