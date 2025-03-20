package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

var db *sql.DB

func SetDB(database *sql.DB) {
	db = database
}

// MovieDetail содержит подробную информацию о фильме
type MovieDetail struct {
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
	Runtime          int     `json:"runtime"`
	Tagline          string  `json:"tagline"`
	Genres           []Genre `json:"genres"`
	Credits          Credits `json:"credits"`
}

// Genre представляет жанр фильма
type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Credits содержит информацию о съемочной группе и актерах
type Credits struct {
	Cast []CastMember `json:"cast"`
}

// CastMember представляет актера
type CastMember struct {
	Name        string `json:"name"`
	Character   string `json:"character"`
	ProfilePath string `json:"profile_path"`
}

// GetMovieDetails получает подробную информацию о фильме по его ID
func GetMovieDetails(movieID string) (*MovieDetail, error) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API_KEY не найден в переменных окружения")
	}

	// Запрос к API TMDB для получения данных о фильме
	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/%s?api_key=%s&language=kk&append_to_response=credits", movieID, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка запроса: %s", resp.Status)
	}

	var movie MovieDetail
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		return nil, err
	}

	return &movie, nil
}

// GetSimilarMovies получает список похожих фильмов
func GetSimilarMovies(movieID string) ([]Movie, error) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API_KEY не найден в переменных окружения")
	}

	// Запрос к API TMDB для получения похожих фильмов
	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/%s/similar?api_key=%s&language=kk", movieID, apiKey)

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
