package services

import (
	"encoding/json" // Для работы с JSON (парсинг ответов от API)
	"fmt"           // Для форматирования строк (например, URL-адресов)
	"net/http"      // Для выполнения HTTP-запросов
	"time"          // Для установки таймаута HTTP-клиента

	"effective-mobile/config" // Импортируем наш логгер
	"effective-mobile/models" // Импортируем нашу модель Person
)

// --- Вспомогательные структуры для парсинга JSON-ответов от внешних API ---

// AgifyResponse представляет структуру ответа от api.agify.io
type AgifyResponse struct {
	Age int `json:"age"` // Поле "age" из JSON-ответа
	// Другие поля в ответе (count, name) нам не нужны, поэтому их здесь нет.
}

// GenderizeResponse представляет структуру ответа от api.genderize.io
type GenderizeResponse struct {
	Gender string `json:"gender"` // Поле "gender" из JSON-ответа
	// Другие поля в ответе (count, name, probability) нам не нужны.
}

// NationalizeResponse представляет структуру ответа от api.nationalize.io
type NationalizeResponse struct {
	Country []struct { // nationalize.io возвращает массив стран
		CountryID   string  `json:"country_id"`   // Код страны (например, "US", "RU")
		Probability float64 `json:"probability"` // Вероятность
	} `json:"country"`
	// Другие поля в ответе (count, name) нам не нужны.
}

// --- Основная функция обогащения данных ---

// EnrichPerson принимает указатель на структуру models.Person
// и пытается обогатить её полями Age, Gender и Nationality,
// обращаясь к внешним API.
func EnrichPerson(person *models.Person) error {
	config.Log.Debugf("Начинаем обогащение данных для человека: %s", person.Name)

	// Создаем HTTP-клиент с таймаутом, чтобы запросы не висли бесконечно.
	client := http.Client{Timeout: 5 * time.Second}

	// --- Обогащение возрастом (agify.io) ---
	ageURL := fmt.Sprintf("https://api.agify.io/?name=%s", person.Name)
	var agifyRes AgifyResponse
	// Вызываем вспомогательную функцию fetchData для выполнения запроса и парсинга ответа.
	if err := fetchData(client, ageURL, &agifyRes); err != nil {
		// Логируем ошибку, но не возвращаем её, так как это не критично.
		config.Log.Warnf("Не удалось обогатить возраст для %s: %v", person.Name, err)
	} else if agifyRes.Age != 0 { // Проверяем, что возраст получен
		person.Age = &agifyRes.Age // Присваиваем указателю на int
		config.Log.Debugf("Обогащен возраст для %s: %d", person.Name, *person.Age)
	}

	// --- Обогащение полом (genderize.io) ---
	genderURL := fmt.Sprintf("https://api.genderize.io/?name=%s", person.Name)
	var genderizeRes GenderizeResponse
	if err := fetchData(client, genderURL, &genderizeRes); err != nil {
		config.Log.Warnf("Не удалось обогатить пол для %s: %v", person.Name, err)
	} else if genderizeRes.Gender != "" { // Проверяем, что пол получен
		person.Gender = &genderizeRes.Gender // Присваиваем указателю на string
		config.Log.Debugf("Обогащен пол для %s: %s", person.Name, *person.Gender)
	}

	// --- Обогащение национальностью (nationalize.io) ---
	nationalityURL := fmt.Sprintf("https://api.nationalize.io/?name=%s", person.Name)
	var nationalizeRes NationalizeResponse
	if err := fetchData(client, nationalityURL, &nationalizeRes); err != nil {
		config.Log.Warnf("Не удалось обогатить национальность для %s: %v", person.Name, err)
	} else if len(nationalizeRes.Country) > 0 {
		// nationalize.io возвращает массив стран с вероятностями.
		// Берем самую вероятную (первую в массиве, если API сортирует по убыванию вероятности, что обычно так).
		person.Nationality = &nationalizeRes.Country[0].CountryID // Присваиваем указателю на string
		config.Log.Debugf("Обогащена национальность для %s: %s", person.Name, *person.Nationality)
	}

	config.Log.Debugf("Завершено обогащение данных для человека: %s", person.Name)
	return nil // Возвращаем nil, если обогащение прошло без критических ошибок.
}

// --- Вспомогательная функция для выполнения HTTP-запросов и парсинга JSON ---

// fetchData выполняет GET-запрос к указанному URL и десериализует JSON-ответ
// в предоставленную целевую структуру (target).
func fetchData(client http.Client, url string, target interface{}) error {
	config.Log.Debugf("Выполнение HTTP-запроса к: %s", url)
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("ошибка при выполнении HTTP-запроса к %s: %w", url, err)
	}
	defer resp.Body.Close() // Гарантируем закрытие тела ответа после использования.

	if resp.StatusCode != http.StatusOK {
		// Если статус-код не 200 OK, значит, что-то пошло не так на стороне API.
		return fmt.Errorf("получен некорректный статус-код от %s: %d", url, resp.StatusCode)
	}

	// Десериализуем (парсим) JSON-ответ в целевую структуру.
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("ошибка при парсинге JSON-ответа от %s: %w", url, err)
	}

	config.Log.Debugf("Успешный ответ от: %s", url)
	return nil
}