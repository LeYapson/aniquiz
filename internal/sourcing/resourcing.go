package sourcing

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/LeYapson/aniquiz/internal/database"
)

// resourcingDelay espace les appels à themes.moe pour ne pas se faire bloquer.
const resourcingDelay = 500 * time.Millisecond

// ResourcingProgress décrit l'avancement du re-sourcing audio.
type ResourcingProgress struct {
	Running    bool      `json:"running"`
	Total      int       `json:"total"`   // nombre de pistes mortes à traiter
	Checked    int       `json:"checked"` // mal_ids interrogés
	Restored   int       `json:"restored"`
	Failed     int       `json:"failed"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
	LastError  string    `json:"last_error,omitempty"`
}

var (
	resourcingMu    sync.Mutex
	resourcingState ResourcingProgress
)

// ResourcingStatus retourne une copie de l'état courant du re-sourcing.
func ResourcingStatus() ResourcingProgress {
	resourcingMu.Lock()
	defer resourcingMu.Unlock()
	return resourcingState
}

// StartAudioResourcing tente de retrouver des URLs valides pour toutes les pistes
// marquées not_found, en rappelant themes.moe par mal_id.
// Retourne false si un re-sourcing tourne déjà.
func StartAudioResourcing() bool {
	resourcingMu.Lock()
	if resourcingState.Running {
		resourcingMu.Unlock()
		return false
	}
	resourcingState = ResourcingProgress{Running: true, StartedAt: time.Now()}
	resourcingMu.Unlock()

	go runAudioResourcing()
	return true
}

func runAudioResourcing() {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("Audio resourcing: panic récupérée: %v", rec)
		}
		resourcingMu.Lock()
		resourcingState.Running = false
		resourcingState.FinishedAt = time.Now()
		final := resourcingState
		resourcingMu.Unlock()
		log.Printf("Audio resourcing terminé: %d restaurées, %d échecs (sur %d pistes mortes)",
			final.Restored, final.Failed, final.Total)
	}()

	refs, err := database.GetDeadTrackRefs()
	if err != nil {
		resourcingMu.Lock()
		resourcingState.LastError = err.Error()
		resourcingMu.Unlock()
		return
	}

	resourcingMu.Lock()
	resourcingState.Total = len(refs)
	resourcingMu.Unlock()

	if len(refs) == 0 {
		return
	}

	// Regroupe les pistes mortes par mal_id pour n'appeler themes.moe qu'une fois
	// par anime, même si plusieurs pistes du même anime sont mortes.
	type malGroup struct {
		tracks []database.DeadTrackRef
	}
	groups := make(map[int]*malGroup)
	order := []int{}
	for _, r := range refs {
		if _, ok := groups[r.MalID]; !ok {
			groups[r.MalID] = &malGroup{}
			order = append(order, r.MalID)
		}
		groups[r.MalID].tracks = append(groups[r.MalID].tracks, r)
	}

	for _, malID := range order {
		group := groups[malID]

		audioLinks, err := GetAudioURL(malID)
		resourcingMu.Lock()
		resourcingState.Checked++
		if err != nil {
			resourcingState.LastError = fmt.Sprintf("mal_id %d: %v", malID, err)
			resourcingState.Failed += len(group.tracks)
			resourcingMu.Unlock()
			time.Sleep(resourcingDelay)
			continue
		}
		resourcingMu.Unlock()

		for _, t := range group.tracks {
			// Reconstruit la clé themes.moe : OP1, OP2, ED1, ED2...
			key := fmt.Sprintf("%s%d", t.TrackType, t.Position)
			newURL, found := audioLinks[key]

			resourcingMu.Lock()
			if found && newURL != "" {
				if updateErr := database.UpdateTrackAudioURL(t.ID, newURL); updateErr == nil {
					resourcingState.Restored++
				} else {
					resourcingState.Failed++
					resourcingState.LastError = updateErr.Error()
				}
			} else {
				resourcingState.Failed++
			}
			resourcingMu.Unlock()
		}

		time.Sleep(resourcingDelay)
	}
}
