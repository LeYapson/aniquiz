package sourcing

import (
	"fmt"
	"github.com/darenliang/jikan-go"
)

func GetAnimeInfo(animeId int) string {
	anime, err := jikan.GetAnimeById(animeId)
	if err != nil {
		panic(err)
	}
	fmt.Println(anime.Data.Title)
	return anime.Data.Title
}