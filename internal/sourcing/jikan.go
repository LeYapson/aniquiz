package sourcing

import (
	"fmt"
	"strings"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/LeYapson/aniquiz/internal/models"
	"github.com/darenliang/jikan-go"
)

type AnimeMusicInfo struct {
	Title     string
	AltTitles []string // titre anglais + synonymes
	Openings  []string
	Endings   []string
	Year      int
}

// buildAltTitles assemble les titres alternatifs (anglais + synonymes), en
// retirant les vides et ceux identiques au titre principal (insensible à la casse).
func buildAltTitles(main, english string, synonyms []string) []string {
	seen := map[string]bool{strings.ToLower(strings.TrimSpace(main)): true}
	var alts []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		key := strings.ToLower(s)
		if s != "" && !seen[key] {
			seen[key] = true
			alts = append(alts, s)
		}
	}
	add(english)
	for _, syn := range synonyms {
		add(syn)
	}
	return alts
}

func ProcessAndSaveAnime(animeId int) (*AnimeMusicInfo, error) {
	// 1. Récupération Jikan
	musicInfo, err := GetAnimeMusic(animeId)
	if err != nil {
		return nil, err
	}

	// 2. Récupération Audio (Linker)
	audiolinks, _ := GetAudioURL(animeId)

	// 3. Openings
	for i, op := range musicInfo.Openings {
		cleanTitle, cleanArtist := parseTrack(op)
		opKey := fmt.Sprintf("OP%d", i+1)
		audioURL := "not_found"
		if url, ok := audiolinks[opKey]; ok {
			audioURL = url
		}
		database.SaveTrack(models.Track{
			Title:      cleanTitle,
			Artist:     cleanArtist,
			AnimeName:  musicInfo.Title,
			AltTitles:  musicInfo.AltTitles,
			AudioURL:   audioURL,
			MalID:      animeId,
			Difficulty: 1,
			TrackType:  "OP",
			AnimeYear:  musicInfo.Year,
		})
	}

	// 4. Endings
	for i, ed := range musicInfo.Endings {
		cleanTitle, cleanArtist := parseTrack(ed)
		edKey := fmt.Sprintf("ED%d", i+1)
		audioURL := "not_found"
		if url, ok := audiolinks[edKey]; ok {
			audioURL = url
		}
		database.SaveTrack(models.Track{
			Title:      cleanTitle,
			Artist:     cleanArtist,
			AnimeName:  musicInfo.Title,
			AltTitles:  musicInfo.AltTitles,
			AudioURL:   audioURL,
			MalID:      animeId,
			Difficulty: 1,
			TrackType:  "ED",
			AnimeYear:  musicInfo.Year,
		})
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
		Title:     anime.Data.Title,
		AltTitles: buildAltTitles(anime.Data.Title, anime.Data.TitleEnglish, anime.Data.TitleSynonyms),
		Openings:  themes.Data.Openings,
		Endings:   themes.Data.Endings,
		Year:      anime.Data.Year,
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
