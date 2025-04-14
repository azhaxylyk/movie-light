package models

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/agnivade/levenshtein"
)

// Расширенная карта жанров с поддержкой множественных жанров для одного слова
var GenreKeywords = map[string][]int{
	// Романтика/Любовь
	"махаббат":      {10749},
	"романтика":     {10749},
	"ғашық":         {10749},
	"сүйіспеншілік": {10749},
	"роман":         {10749, 18}, // может быть и романом, и драмой

	// Ужасы
	"қорқыныш": {27},
	"ужас":     {27},
	"қорқақ":   {27},

	// Приключения
	"оқиға":       {12},
	"шытырман":    {12},
	"приключение": {12},
	"серіктестік": {12},
	"саяхат":      {12, 28}, // может быть и приключением, и экшном

	// Боевики
	"экшн":   {28},
	"боевик": {28},
	"шайқас": {28},
	"ұрыс":   {28},
	"жанжал": {28},
	"соғыс":  {28, 18}, // может быть и боевиком, и драмой
	"атыс":   {28},

	// Комедии
	"көңілді": {35},
	"комедия": {35},
	"күлімді": {35},
	"әзіл":    {35},
	"сатира":  {35, 18}, // может быть и комедией, и драмой

	// Драмы
	"драма":    {18},
	"қайғы":    {18},
	"эмоция":   {18},
	"трагедия": {18},
	"арман":    {18},
	"өмірбаян": {18},
	"тағдыр":   {18},

	// Мультфильмы
	"мультфильм": {16},
	"анимация":   {16},
	"мультик":    {16},
	"балалар":    {16},
	"қарқара":    {16},

	// Фантастика
	"фантастика": {878},
	"ғылыми":     {878},
	"болашақ":    {878},
	"ғарыш":      {878},
	"планета":    {878},

	// Фэнтези
	"фэнтези": {14},
	"сиқыр":   {14},
	"қиял":    {14},
	"аңыз":    {14},
}

// Список популярных жанров для подсказки
var PopularGenres = []string{
	"комедия", "романтика", "фантастика", "фэнтези",
	"боевик", "драма", "қорқынышты", "мультфильм",
}

// DetectGenres определяет жанры с улучшенной обработкой
func DetectGenres(text string) ([]int, string) {
	text = normalizeKazakhText(strings.ToLower(text))
	detectedGenres := make(map[int]bool)
	foundAny := false

	// 1. Проверка точных совпадений
	for keyword, genreIDs := range GenreKeywords {
		if strings.Contains(text, keyword) {
			foundAny = true
			for _, genreID := range genreIDs {
				detectedGenres[genreID] = true
			}
		}
	}

	// 2. Морфологический анализ и нечеткий поиск
	if !foundAny {
		words := strings.Fields(text)
		for _, word := range words {
			baseWord := removeKazakhSuffix(word)

			for keyword, genreIDs := range GenreKeywords {
				baseKeyword := removeKazakhSuffix(keyword)

				// Проверяем похожесть слов
				if len(baseWord) > 3 && len(baseKeyword) > 3 {
					distance := levenshtein.ComputeDistance(baseWord, baseKeyword)
					if distance <= 2 {
						for _, genreID := range genreIDs {
							detectedGenres[genreID] = true
						}
					}
				}
			}
		}
	}

	// Конвертируем в slice
	var result []int
	for genreID := range detectedGenres {
		result = append(result, genreID)
	}

	// Генерация подсказки, если ничего не найдено
	var hint string
	if len(result) == 0 {
		hint = generateGenreHint()
	}

	return result, hint
}

// Генерация подсказки по жанрам
func generateGenreHint() string {
	return fmt.Sprintf("Қай жанрды қалайсыз? Мысалы: %s, ...",
		strings.Join(PopularGenres, ", "))
}

// Удаление казахских суффиксов
func removeKazakhSuffix(word string) string {
	suffixes := []string{
		"тық", "тік", "қа", "ке", "тан", "тен",
		"пен", "мен", "бен", "менен",
		"дың", "дің", "ның", "нің",
		"дар", "дер", "тар", "тер",
		"ты", "ті", "лы", "лі",
	}

	for _, s := range suffixes {
		if strings.HasSuffix(word, s) && len(word) > len(s)+2 {
			return word[:len(word)-len(s)]
		}
	}
	return word
}

// Нормализация казахского текста
func normalizeKazakhText(text string) string {
	var result strings.Builder
	for _, r := range text {
		switch r {
		case 'ә':
			result.WriteString("а")
		case 'ғ':
			result.WriteString("г")
		case 'қ':
			result.WriteString("к")
		case 'ң':
			result.WriteString("н")
		case 'ө':
			result.WriteString("о")
		case 'ұ':
			result.WriteString("у")
		case 'ү':
			result.WriteString("у")
		case 'һ':
			result.WriteString("х")
		case 'і':
			result.WriteString("и")
		default:
			if unicode.IsLetter(r) || unicode.IsNumber(r) {
				result.WriteRune(r)
			}
		}
	}
	return result.String()
}

// ChatResponseByGenre получает фильмы по жанру и формирует текстовый ответ
func ChatResponseByGenre(genreID int) (string, error) {
	movies, err := GetGenreFilms([]int{genreID})
	if err != nil {
		return "", err
	}

	if len(movies) == 0 {
		return "Бұл жанр бойынша фильмдер табылмады.", nil
	}

	// Сформируем краткий текст с предложениями фильмов
	response := fmt.Sprintf("Ұсынылған жанр бойынша фильмдер:\n")
	for i, movie := range movies {
		if i >= 5 { // ограничим до 5 фильмов
			break
		}
		response += fmt.Sprintf("- %s (%s)\n", movie.Title, movie.ReleaseDate)
	}

	return response, nil
}
