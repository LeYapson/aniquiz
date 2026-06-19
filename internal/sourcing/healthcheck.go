package sourcing

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/LeYapson/aniquiz/internal/database"
)

// audioCheckDelay espace les requêtes pour ne pas marteler l'hôte des mirrors.
const audioCheckDelay = 200 * time.Millisecond

// healthClient : timeout court, on ne télécharge pas le fichier entier.
var healthClient = &http.Client{Timeout: 10 * time.Second}

// AudioCheckProgress décrit l'avancement de la vérification des liens audio.
type AudioCheckProgress struct {
	Running     bool      `json:"running"`
	Total       int       `json:"total"`
	Checked     int       `json:"checked"`
	Dead        int       `json:"dead"`        // liens morts confirmés → marqués not_found
	Unreachable int       `json:"unreachable"` // injoignables (transitoire) → NON modifiés
	StartedAt   time.Time `json:"started_at"`
	FinishedAt  time.Time `json:"finished_at,omitempty"`
	LastError   string    `json:"last_error,omitempty"`
}

var (
	checkMu    sync.Mutex
	checkState AudioCheckProgress
)

// AudioCheckStatus retourne une copie de l'état courant de la vérification.
func AudioCheckStatus() AudioCheckProgress {
	checkMu.Lock()
	defer checkMu.Unlock()
	return checkState
}

// StartAudioHealthcheck lance, en arrière-plan, la vérification de toutes les
// URLs audio. Les liens définitivement morts (404/410) sont marqués 'not_found'
// pour être exclus du jeu ; les erreurs transitoires (timeout, réseau) sont
// laissées intactes afin de ne pas détruire un lien valable sur un simple aléa.
// Retourne false si une vérification tourne déjà.
func StartAudioHealthcheck() bool {
	checkMu.Lock()
	if checkState.Running {
		checkMu.Unlock()
		return false
	}
	checkState = AudioCheckProgress{Running: true, StartedAt: time.Now()}
	checkMu.Unlock()

	go runAudioHealthcheck()
	return true
}

func runAudioHealthcheck() {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("Audio healthcheck: panic récupérée: %v", rec)
		}
		checkMu.Lock()
		checkState.Running = false
		checkState.FinishedAt = time.Now()
		final := checkState
		checkMu.Unlock()
		log.Printf("Audio healthcheck terminé: %d vérifiés, %d morts, %d injoignables",
			final.Checked, final.Dead, final.Unreachable)
	}()

	refs, err := database.GetTracksWithAudio()
	if err != nil {
		checkMu.Lock()
		checkState.LastError = err.Error()
		checkMu.Unlock()
		return
	}

	checkMu.Lock()
	checkState.Total = len(refs)
	checkMu.Unlock()

	for _, ref := range refs {
		switch checkAudio(ref.AudioURL) {
		case audioDead:
			checkMu.Lock()
			checkState.Checked++
			if err := database.MarkAudioNotFound(ref.ID); err == nil {
				checkState.Dead++
			} else {
				checkState.LastError = err.Error()
			}
			checkMu.Unlock()
		case audioUnreachable:
			checkMu.Lock()
			checkState.Checked++
			checkState.Unreachable++
			checkMu.Unlock()
		default: // audioOK
			checkMu.Lock()
			checkState.Checked++
			checkMu.Unlock()
		}
		time.Sleep(audioCheckDelay)
	}
}

type audioStatus int

const (
	audioOK audioStatus = iota
	audioDead
	audioUnreachable
)

// checkAudio teste la disponibilité d'une URL audio sans la télécharger.
// HEAD d'abord ; si l'hôte refuse HEAD, repli sur un GET Range 0-0.
func checkAudio(url string) audioStatus {
	if s := classify(http.MethodHead, url, false); s != audioUnreachable {
		// HEAD a tranché (OK ou mort), sauf si l'hôte ne supporte pas HEAD.
		return s
	}
	// Repli GET partiel (certains hôtes renvoient 405/403 sur HEAD).
	return classify(http.MethodGet, url, true)
}

func classify(method, url string, rangeReq bool) audioStatus {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return audioDead // URL malformée : irrécupérable
	}
	if rangeReq {
		req.Header.Set("Range", "bytes=0-0")
	}
	resp, err := healthClient.Do(req)
	if err != nil {
		return audioUnreachable // timeout / DNS / réseau : transitoire
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 400:
		return audioOK
	case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
		return audioDead
	case method == http.MethodHead && (resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusForbidden):
		return audioUnreachable // pousse vers le repli GET
	default:
		return audioUnreachable // 5xx, etc. : on ne détruit pas le lien
	}
}
