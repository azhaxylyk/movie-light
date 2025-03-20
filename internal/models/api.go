package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

var apiKey string

type Movie struct {
	Title      string `json:"title"`
	Overview   string `json:"overview"`
	PosterPath string `json:"poster_path"`
}

func InitAPI() {
	apiKey = os.Getenv("API_KEY")
}

func GetPopularMovies() ([]Movie, error) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API_KEY не найден в переменных окружения")
	}

	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/popular?api_key=%s&language=kk", apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ошибка запроса: %s", resp.Status)
	}

	var data struct {
		Results []Movie `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Results, nil
}
