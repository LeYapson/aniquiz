package sourcing

import (
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
	//1 - On recupere les infos de l'anime via jikan 
	musicInfo, err := GetAnimeMusic(animeId)
	if err != nil {
		return nil, err
	}

	//2 - Pour chaque opening trouvé, on crée un modele et on sauvegarde
	for _, op := range musicInfo.Openings {
		track := models.Track{
			Title: op, // On pourrait faire mieux en essayant d'extraire le titre de la chanson
			AnimeName: musicInfo.Title,
			AudioURL: "pending", // On remplira ça plus tard
			Difficulty: 1, // Valeur par défaut pour l'instant
		}

		//3 - On sauvegarde dans la base de données
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