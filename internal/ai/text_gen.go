// internal/ai/text_gen.go
package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ai-content-gen/internal/config" // Импортируем конфиг
	"ai-content-gen/pkg/utils"
)

// ResponseChoice представляет элемент в массиве choices ответа AI.
type ResponseChoice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

// OpenAICompletionResponse соответствует структуре ответа от локальной модели.
type OpenAICompletionResponse struct {
	Choices []ResponseChoice `json:"choices"`
}

// TextGenerator представляет интерфейс для текстовой нейросети.
type TextGenerator struct {
	Endpoint string
	Config   *config.AppConfig // Ссылка на AppConfig
	Logger   *utils.Logger
}

// NewTextGenerator создает новый экземпляр TextGenerator.
func NewTextGenerator(endpoint string, cfg *config.AppConfig, logger *utils.Logger) *TextGenerator {
	return &TextGenerator{
		Endpoint: endpoint,
		Config:   cfg,
		Logger:   logger,
	}
}

// GenerateShortsIdeaAndScenes генерирует общую идею и краткое описание сцен для YouTube Shorts.
func (tg *TextGenerator) GenerateShortsIdeaAndScenes(topic string) (string, error) {
	tg.Logger.Info("Запрос на генерацию общей идеи и сцен для темы: %s", topic)

	promptContent := fmt.Sprintf(`Придумай идею для YouTube Shorts про "%s".
Формат ответа строго следующий:
Идея: [краткое описание идеи]

Сцена 1: [краткое описание]
Сцена 2: [краткое описание]
Сцена 3: [краткое описание]
... (до 5-7 сцен, если уместно)
`, topic)

	// Используем max_tokens_general из конфигурации
	return tg.callAI(promptContent, tg.Config.AI.Text.MaxTokensGeneral)
}

// GenerateVideoPromptForScene генерирует подробный промпт для видеогенерации конкретной сцены.
func (tg *TextGenerator) GenerateVideoPromptForScene(overallIdea, sceneDescription string) (string, error) {
	tg.Logger.Info("Запрос на генерацию подробного промпта для видео-сцены: %s", sceneDescription)

	promptContent := fmt.Sprintf(`На основе общей идеи "%s" и описания сцены "%s",
создай очень подробный и детализированный промпт, пригодный для прямой генерации видео.
Опиши: что происходит в кадре, какие объекты присутствуют, их действия, фон, освещение, настроение, стиль.
Сфокусируйся на визуальных деталях.
`, overallIdea, sceneDescription)

	// Используем max_tokens_detailed из конфигурации
	return tg.callAI(promptContent, tg.Config.AI.Text.MaxTokensDetailed)
}

// callAI является внутренней функцией для отправки запросов к локальной модели.
func (tg *TextGenerator) callAI(content string, maxTokens int) (string, error) {
	requestBody := map[string]interface{}{
		"model":                tg.Config.AI.Text.Model, // Модель из YAML
		"chat_template_kwargs": map[string]bool{"enable_thinking": false},
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": content,
			},
		},
		"max_tokens":  maxTokens,                     // Динамический max_tokens
		"temperature": tg.Config.AI.Text.Temperature, // Температура из YAML
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("ошибка при маршалинге JSON запроса: %w", err)
	}

	resp, err := http.Post(tg.Endpoint, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("ошибка при отправке запроса к текстовой нейросети: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("получен некорректный статус от текстовой нейросети: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка при чтении ответа от текстовой нейросети: %w", err)
	}

	var responseData OpenAICompletionResponse
	err = json.Unmarshal(bodyBytes, &responseData)
	if err != nil {
		return "", fmt.Errorf("ошибка при демаршалинге JSON ответа: %w\nОтвет: %s", err, string(bodyBytes))
	}

	if len(responseData.Choices) == 0 {
		return "", fmt.Errorf("не найдено 'choices' в ответе от текстовой нейросети")
	}

	return responseData.Choices[0].Message.Content, nil
}
