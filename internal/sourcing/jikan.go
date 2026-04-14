package sourcing

import (
	"github.com/darenliang/jikan-go"
)

type AnimeMusicInfo struct {
	Title string
	Openings []string
	Endings []string

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