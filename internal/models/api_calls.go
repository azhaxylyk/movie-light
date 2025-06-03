package models

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

var apiKey string
var YouTubeAPIKey string

const baseURL = "https://api.themoviedb.org/3/"

func InitAPI() {
	apiKey = os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY не найден в переменных окружения")
	}
	YouTubeAPIKey = os.Getenv("YOU_TUBE_API_KEY")
	if YouTubeAPIKey == "" {
		log.Fatal("YouTubeAPIKey не найден в переменных окружения")
	}
}

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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
	Videos           struct {
		Results []struct {
			Key  string
			Site string
			Type string
			Name string
		}
	}
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

func FetchFromAPI(endpoint string, target interface{}) error {
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
	if err := FetchFromAPI("movie/top_rated", &data); err != nil {
		return nil, err
	}
	log.Printf("Got %d movies from API", len(data.Results))
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
	if err := FetchFromAPI(fmt.Sprintf("trending/movie/%s", timeWindow), &data); err != nil {
		return nil, err
	}
	return data.Results, nil
}

// GetMovieDetails получает подробную информацию о фильме по его ID
func GetMovieDetails(movieID string) (*MovieDetail, error) {
	var movie MovieDetail
	endpoint := fmt.Sprintf("movie/%s?append_to_response=credits,videos,images", movieID)
	if err := FetchFromAPI(endpoint, &movie); err != nil {
		return nil, err
	}
	return &movie, nil
}

// GetSimilarMovies получает список похожих фильмов
func GetSimilarMovies(movieID string) ([]Movie, error) {
	var data struct {
		Results []Movie `json:"results"`
	}
	if err := FetchFromAPI(fmt.Sprintf("movie/%s/similar", movieID), &data); err != nil {
		return nil, err
	}
	return data.Results, nil
}

// SearchMovies выполняет поиск фильмов с фильтрами
func SearchMovies(query string, filters map[string]string) ([]Movie, error) {
	var data struct {
		Results []Movie `json:"results"`
	}

	// Создаем базовый endpoint
	endpoint := fmt.Sprintf("search/movie?query=%s", url.QueryEscape(query))

	// Добавляем фильтры к запросу
	for key, value := range filters {
		if value != "" {
			endpoint += fmt.Sprintf("&%s=%s", key, url.QueryEscape(value))
		}
	}

	if err := FetchFromAPI(endpoint, &data); err != nil {
		return nil, err
	}
	return data.Results, nil
}

// getGenres получает список жанров фильмов
func GetGenres() ([]Genre, error) {
	var response struct {
		Genres []Genre `json:"genres"`
	}

	// Используем существующую функцию fetchFromAPI из вашего models пакета
	err := FetchFromAPI("genre/movie/list", &response)
	if err != nil {
		log.Printf("Ошибка при получении жанров: %v", err)
		return nil, fmt.Errorf("не удалось получить список жанров")
	}

	return response.Genres, nil
}

func GetGenreFilms(genres []int) ([]Movie, error) {
	if len(genres) == 0 {
		log.Println("Список жанров пуст, возвращаем популярные фильмы")
		return GetPopularMovies()
	}

	// Преобразуем массив жанров в строку, разделённую запятыми
	genreIDs := ""
	for i, id := range genres {
		if i > 0 {
			genreIDs += ","
		}
		genreIDs += fmt.Sprintf("%d", id)
	}

	endpoint := fmt.Sprintf("discover/movie?with_genres=%s&sort_by=popularity.desc", url.QueryEscape(genreIDs))

	var data struct {
		Results []Movie `json:"results"`
	}

	if err := FetchFromAPI(endpoint, &data); err != nil {
		log.Printf("Ошибка при получении фильмов по жанрам: %v", err)
		return nil, err
	}

	return data.Results, nil
}

func GetMovieTrailer(movieID int) (string, error) {
	var trailerData struct {
		Results []struct {
			Key  string `json:"key"`
			Type string `json:"type"`
			Site string `json:"site"`
		} `json:"results"`
	}

	endpoint := fmt.Sprintf("%smovie/%d/videos?api_key=%s", baseURL, movieID, apiKey)
	resp, err := http.Get(endpoint)
	if err != nil {
		return "", fmt.Errorf("ошибка запроса к TMDB API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("TMDB API вернул статус %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	if err := json.Unmarshal(body, &trailerData); err != nil {
		return "", fmt.Errorf("ошибка парсинга JSON: %v", err)
	}

	for _, video := range trailerData.Results {
		if video.Type == "Trailer" && video.Site == "YouTube" {
			return fmt.Sprintf("https://www.youtube.com/watch?v=%s", video.Key), nil
		}
	}

	return "", fmt.Errorf("трейлер для фильма с ID %d не найден", movieID)
}
