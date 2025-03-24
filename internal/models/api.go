package models

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

var apiKey string

func InitAPI() {
	apiKey = os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY не найден в переменных окружения")
	}
}

// Movie представляет краткую информацию о фильме
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

func fetchFromAPI(endpoint string, target interface{}) error {
	if apiKey == "" {
		return fmt.Errorf("API_KEY не установлен")
	}

	baseURL := fmt.Sprintf("https://api.themoviedb.org/3/%s", endpoint)

	// Парсим URL для правильного добавления параметров
	u, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

	// Получаем текущие параметры запроса
	q := u.Query()

	// Добавляем обязательные параметры, если их еще нет
	if q.Get("api_key") == "" {
		q.Add("api_key", apiKey)
	}
	if q.Get("language") == "" {
		q.Add("language", "ru")
	}

	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		log.Printf("Ошибка при запросе к API: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка запроса: %s, тело ответа: %s", resp.Status, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		log.Printf("Ошибка декодирования JSON: %v", err)
		return err
	}

	return nil
}

// GetPopularMovies получает список популярных фильмов
func GetPopularMovies() ([]Movie, error) {
	var data struct {
		Results []Movie `json:"results"`
	}
	if err := fetchFromAPI("movie/top_rated", &data); err != nil {
		return nil, err
	}
	return data.Results, nil
}

// GetTrendingMovies получает список трендовых фильмов за день или неделю
func GetTrendingMovies(timeWindow string) ([]Movie, error) {
	if timeWindow != "day" && timeWindow != "week" {
		return nil, fmt.Errorf("неверный временной интервал. Используйте 'day' или 'week'")
	}

	var data struct {
		Results []Movie `json:"results"`
	}
	if err := fetchFromAPI(fmt.Sprintf("trending/movie/%s", timeWindow), &data); err != nil {
		return nil, err
	}
	return data.Results, nil
}

// GetMovieDetails получает подробную информацию о фильме по его ID
func GetMovieDetails(movieID string) (*MovieDetail, error) {
	var movie MovieDetail
	endpoint := fmt.Sprintf("movie/%s?append_to_response=credits,videos,images", movieID)
	if err := fetchFromAPI(endpoint, &movie); err != nil {
		return nil, err
	}
	return &movie, nil
}

// GetSimilarMovies получает список похожих фильмов
func GetSimilarMovies(movieID string) ([]Movie, error) {
	var data struct {
		Results []Movie `json:"results"`
	}
	if err := fetchFromAPI(fmt.Sprintf("movie/%s/similar", movieID), &data); err != nil {
		return nil, err
	}
	return data.Results, nil
}
