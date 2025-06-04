package models

import (
	"fmt"
	"log"
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
	"қорқыныш":   {27},
	"қорқынышты": {27},
	"ужас":       {27},
	"қорқақ":     {27},
	"хоррор":     {27},

	// Приключения
	"оқиға":       {12},
	"оқиғалы":     {12},
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
	"күлкілі": {35},
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

// Улучшенная функция DetectGenres
func DetectGenres(text string) ([]int, string) {
	// Сначала нормализуем текст (сохраняя пробелы)
	normalizedText := normalizeKazakhText(strings.ToLower(text))
	log.Printf("Нормализованный текст: '%s'", normalizedText)

	detectedGenres := make(map[int]bool)
	foundAny := false

	// 1. Разбиваем на слова с учетом казахских особенностей
	words := strings.Fields(normalizedText)

	// 2. Проверяем каждое слово
	for _, word := range words {
		// Пробуем найти точное совпадение
		if genreIDs, ok := GenreKeywords[word]; ok {
			foundAny = true
			for _, genreID := range genreIDs {
				detectedGenres[genreID] = true
			}
			continue
		}

		// Если точного совпадения нет, пробуем удалить суффиксы
		baseWord := removeKazakhSuffix(word)
		if baseWord != word {
			if genreIDs, ok := GenreKeywords[baseWord]; ok {
				foundAny = true
				for _, genreID := range genreIDs {
					detectedGenres[genreID] = true
				}
				continue
			}
		}

		// Нечеткий поиск для слов длиной > 3
		if len(baseWord) > 3 {
			for keyword, genreIDs := range GenreKeywords {
				baseKeyword := removeKazakhSuffix(keyword)
				distance := levenshtein.ComputeDistance(baseWord, baseKeyword)
				if distance <= 2 {
					foundAny = true
					for _, genreID := range genreIDs {
						detectedGenres[genreID] = true
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

	// Генерация подсказки
	var hint string
	if !foundAny {
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
		"ты", "ті", "ды", "ді", "ны", "ні", // падежные окончания
		"лық", "лік", "дық", "дік", "тық", "тік", // образует прилагательные
		"шы", "ші", // профессии/характеристики
		"да", "де", "та", "те", // местный падеж
		"нан", "нен", "дан", "ден", // исходный падеж
		"мен", "бен", "пен", // инструментальный падеж
		"ша", "ше", // сравнительная степень
		"ға", "ге", "қа", "ке", // направительный падеж
		"ар", "ер", "йыр", // образует глаголы
		"у", "ү", "ы", "і", // инфинитив
	}

	for _, s := range suffixes {
		if strings.HasSuffix(word, s) && len(word) > len(s)+2 {
			return word[:len(word)-len(s)]
		}
	}
	return word
}

// Нормализация казахского текста с сохранением пробелов
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
		case ' ':
			result.WriteRune(' ') // Сохраняем пробелы
		default:
			if unicode.IsLetter(r) || unicode.IsNumber(r) {
				result.WriteRune(r)
			}
		}
	}
	return result.String()
}
