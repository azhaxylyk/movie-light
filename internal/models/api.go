package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

var apiKey string

type Movie struct {
	ID               int     `json:"id"`
	Title            string  `json:"title"`
	Overview         string  `json:"overview"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	OriginalTitle    string  `json:"original_title"`
	OriginalLanguage string  `json:"original_language"`
	Popularity       float64 `json:"popularity"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
	GenreIDs         []int   `json:"genre_ids"`
	Video            bool    `json:"video"`
}

func InitAPI() {
	apiKey = os.Getenv("API_KEY")
}

func GetPopularMovies() ([]Movie, error) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API_KEY не найден в переменных окружения")
	}

	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/top_rated?api_key=%s&language=ru", apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка запроса: %s", resp.Status)
	}

	var data struct {
		Results []Movie `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Results, nil
}

func GetTrendingMovies(timeWindow string) ([]Movie, error) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API_KEY не найден в переменных окружения")
	}

	if timeWindow != "day" && timeWindow != "week" {
		return nil, fmt.Errorf("неверный временной интервал. Используйте 'day' или 'week'")
	}

	url := fmt.Sprintf("https://api.themoviedb.org/3/trending/movie/%s?api_key=%s&language=ru", timeWindow, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка запроса: %s", resp.Status)
	}

	var data struct {
		Results []Movie `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Results, nil
}
