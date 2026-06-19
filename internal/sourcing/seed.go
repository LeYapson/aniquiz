package sourcing

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/LeYapson/aniquiz/internal/database"
	jikan "github.com/darenliang/jikan-go"
)

// jikanDelay respecte la limite de l'API Jikan (~3 req/s). On reste prudent à
// un appel par ~1,2 s entre chaque anime pour éviter les 429.
const jikanDelay = 1200 * time.Millisecond

// animesPerPage : Jikan renvoie 25 entrées par page de "top anime".
const animesPerPage = 25

// maxSeedPages : garde-fou (40 pages ≈ 1000 animes les plus populaires).
const maxSeedPages = 40

// SeedProgress décrit l'avancement du job d'import en masse.
type SeedProgress struct {
	Running    bool      `json:"running"`
	Pages      int       `json:"pages"`
	Total      int       `json:"total"` // estimation : pages * 25
	Processed  int       `json:"processed"`
	Imported   int       `json:"imported"`
	Skipped    int       `json:"skipped"`
	Failed     int       `json:"failed"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
	LastError  string    `json:"last_error,omitempty"`
}

var (
	seedMu    sync.Mutex
	seedState SeedProgress
)

// SeedStatus retourne une copie de l'état courant du job de seed.
func SeedStatus() SeedProgress {
	seedMu.Lock()
	defer seedMu.Unlock()
	return seedState
}

// StartSeed lance, en arrière-plan, l'import des animes les plus populaires de
// MyAnimeList (page par page, 25 animes/page). Importer par popularité maximise
// le recouvrement avec les listes perso AniList/MAL des joueurs.
// Retourne false si un job tourne déjà.
func StartSeed(pages int) bool {
	if pages < 1 {
		pages = 1
	}
	if pages > maxSeedPages {
		pages = maxSeedPages
	}

	seedMu.Lock()
	if seedState.Running {
		seedMu.Unlock()
		return false
	}
	seedState = SeedProgress{
		Running:   true,
		Pages:     pages,
		Total:     pages * animesPerPage,
		StartedAt: time.Now(),
	}
	seedMu.Unlock()

	go runSeed(pages)
	return true
}

func runSeed(pages int) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("Seed: panic récupérée: %v", rec)
		}
		seedMu.Lock()
		seedState.Running = false
		seedState.FinishedAt = time.Now()
		final := seedState
		seedMu.Unlock()
		log.Printf("Seed terminé: %d importés, %d ignorés, %d échecs (sur %d traités)",
			final.Imported, final.Skipped, final.Failed, final.Processed)
	}()

	for page := 1; page <= pages; page++ {
		top, err := jikan.GetTopAnime("", jikan.TopAnimeFilterByPopularity, page)
		time.Sleep(jikanDelay)
		if err != nil {
			seedMu.Lock()
			seedState.LastError = fmt.Sprintf("page %d: %v", page, err)
			seedMu.Unlock()
			continue
		}

		for _, a := range top.Data {
			malID := a.MalId
			if malID == 0 {
				continue
			}

			// Évite de re-télécharger un anime déjà présent.
			if already, err := database.IsAnimeImported(malID); err == nil && already {
				seedMu.Lock()
				seedState.Processed++
				seedState.Skipped++
				seedMu.Unlock()
				continue
			}

			_, err := ProcessAndSaveAnime(malID)

			seedMu.Lock()
			seedState.Processed++
			if err != nil {
				seedState.Failed++
				seedState.LastError = fmt.Sprintf("anime %d: %v", malID, err)
			} else {
				seedState.Imported++
			}
			seedMu.Unlock()

			// ProcessAndSaveAnime fait plusieurs appels Jikan : on espace.
			time.Sleep(jikanDelay)
		}
	}
}
