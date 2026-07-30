package sourcing

import (
	"log"
	"time"
)

// StartDailyScheduler lance en arrière-plan les tâches nocturnes récurrentes :
// 1. healthcheck audio (marque les liens morts)
// 2. re-sourcing audio (retrouve de nouvelles URLs pour les pistes mortes)
// Appeler une fois au démarrage du serveur.
func StartDailyScheduler() {
	go func() {
		for {
			next := nextRunAt(3, 0) // 3h00 UTC chaque nuit
			log.Printf("Scheduler: prochain cycle audio à %s", next.Format(time.RFC3339))
			time.Sleep(time.Until(next))

			// Étape 1 : healthcheck
			log.Println("Scheduler: démarrage du healthcheck audio")
			if StartAudioHealthcheck() {
				// Attendre la fin du healthcheck avant de re-sourcer.
				for AudioCheckStatus().Running {
					time.Sleep(10 * time.Second)
				}
				status := AudioCheckStatus()
				log.Printf("Scheduler: healthcheck terminé (%d morts, %d injoignables)", status.Dead, status.Unreachable)
			} else {
				log.Println("Scheduler: healthcheck déjà en cours, re-sourcing ignoré ce cycle")
				continue
			}

			// Étape 2 : re-sourcing des pistes mortes
			log.Println("Scheduler: démarrage du re-sourcing audio")
			if StartAudioResourcing() {
				for ResourcingStatus().Running {
					time.Sleep(10 * time.Second)
				}
				status := ResourcingStatus()
				log.Printf("Scheduler: re-sourcing terminé (%d restaurées, %d échecs)", status.Restored, status.Failed)
			}
		}
	}()
}

// nextRunAt retourne le prochain instant UTC à l'heure et minute donnés.
// Si ce moment est déjà passé aujourd'hui, retourne le lendemain.
func nextRunAt(hour, minute int) time.Time {
	now := time.Now().UTC()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.UTC)
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}
