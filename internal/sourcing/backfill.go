package sourcing

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/LeYapson/aniquiz/internal/database"
	jikan "github.com/darenliang/jikan-go"
)

// TitleBackfillProgress décrit l'avancement du backfill des titres alternatifs.
type TitleBackfillProgress struct {
	Running    bool      `json:"running"`
	Total      int       `json:"total"`
	Processed  int       `json:"processed"`
	Updated    int       `json:"updated"`
	Failed     int       `json:"failed"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
	LastError  string    `json:"last_error,omitempty"`
}

var (
	backfillMu    sync.Mutex
	backfillState TitleBackfillProgress
)

// TitleBackfillStatus retourne une copie de l'état courant.
func TitleBackfillStatus() TitleBackfillProgress {
	backfillMu.Lock()
	defer backfillMu.Unlock()
	return backfillState
}

// StartTitleBackfill recharge, pour chaque anime déjà en base, ses titres
// alternatifs (anglais + synonymes) depuis Jikan. Nécessaire car le seed ignore
// les animes déjà importés et ne peut donc pas les rétro-remplir.
// Retourne false si un backfill tourne déjà.
func StartTitleBackfill() bool {
	backfillMu.Lock()
	if backfillState.Running {
		backfillMu.Unlock()
		return false
	}
	backfillState = TitleBackfillProgress{Running: true, StartedAt: time.Now()}
	backfillMu.Unlock()

	go runTitleBackfill()
	return true
}

func runTitleBackfill() {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("Backfill titres: panic récupérée: %v", rec)
		}
		backfillMu.Lock()
		backfillState.Running = false
		backfillState.FinishedAt = time.Now()
		final := backfillState
		backfillMu.Unlock()
		log.Printf("Backfill titres terminé: %d mis à jour, %d échecs (sur %d traités)",
			final.Updated, final.Failed, final.Processed)
	}()

	ids, err := database.GetDistinctMalIDs()
	if err != nil {
		backfillMu.Lock()
		backfillState.LastError = err.Error()
		backfillMu.Unlock()
		return
	}

	backfillMu.Lock()
	backfillState.Total = len(ids)
	backfillMu.Unlock()

	for _, malID := range ids {
		anime, err := jikan.GetAnimeById(malID)
		time.Sleep(jikanDelay) // respecte la limite Jikan

		backfillMu.Lock()
		backfillState.Processed++
		if err != nil {
			backfillState.Failed++
			backfillState.LastError = fmt.Sprintf("anime %d: %v", malID, err)
			backfillMu.Unlock()
			continue
		}
		alts := buildAltTitles(anime.Data.Title, anime.Data.TitleEnglish, anime.Data.TitleSynonyms)
		if err := database.UpdateAnimeTitles(malID, alts); err != nil {
			backfillState.Failed++
			backfillState.LastError = err.Error()
		} else {
			backfillState.Updated++
		}
		backfillMu.Unlock()
	}
}
