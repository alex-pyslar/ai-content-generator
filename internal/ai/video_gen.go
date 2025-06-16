// internal/ai/video_gen.go
package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"ai-content-gen/internal/config"
	"ai-content-gen/pkg/utils"
)

// VideoGenerator представляет интерфейс для видео нейросети.
type VideoGenerator struct {
	Endpoint string
	APIKey   string
	Config   *config.AppConfig // Ссылка на AppConfig
	Logger   *utils.Logger
}

// NewVideoGenerator создает новый экземпляр VideoGenerator.
func NewVideoGenerator(endpoint, apiKey string, cfg *config.AppConfig, logger *utils.Logger) *VideoGenerator {
	return &VideoGenerator{
		Endpoint: endpoint,
		APIKey:   apiKey,
		Config:   cfg,
		Logger:   logger,
	}
}

// VideoGenerationRequest соответствует структуре запроса к вашей видео-нейросети.
type VideoGenerationRequest struct {
	Prompt       string `json:"prompt"`
	Resolution   string `json:"resolution"`
	OutputFormat string `json:"output_format"`
	FPS          int    `json:"fps"`
	// Добавьте другие параметры, если ваша модель их поддерживает (например, duration, seed)
}

// VideoGenerationResponse соответствует структуре ответа от вашей видео-нейросети.
type VideoGenerationResponse struct {
	VideoURL string `json:"video_url"` // Предполагаем, что модель возвращает URL видео
	// Или base64 кодированное видео, или путь к файлу на сервере
}

// GenerateVideoSegment генерирует короткий видеофрагмент на основе заданного промпта.
// Возвращает путь к сгенерированному видеофайлу.
func (vg *VideoGenerator) GenerateVideoSegment(prompt string, segmentIndex int) (string, error) {
	vg.Logger.Info("Запрос на генерацию видеофрагмента для промпта (сцена %d): %s", segmentIndex, prompt)

	requestBody := VideoGenerationRequest{
		Prompt:       prompt,
		Resolution:   vg.Config.AI.Video.Resolution,
		OutputFormat: vg.Config.AI.Video.OutputFormat,
		FPS:          vg.Config.AI.Video.FPS,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("ошибка при маршалинге JSON запроса для видео: %w", err)
	}

	req, err := http.NewRequest("POST", vg.Endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("ошибка создания HTTP запроса для видео: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+vg.APIKey) // Если требуется аутентификация

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка при отправке запроса к видео нейросети: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("получен некорректный статус от видео нейросети: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка при чтении ответа от видео нейросети: %w", err)
	}

	var responseData VideoGenerationResponse
	err = json.Unmarshal(bodyBytes, &responseData)
	if err != nil {
		return "", fmt.Errorf("ошибка при демаршалинге JSON ответа видео: %w\nОтвет: %s", err, string(bodyBytes))
	}

	if responseData.VideoURL == "" {
		return "", fmt.Errorf("видео URL не найден в ответе от видео нейросети")
	}

	// Скачиваем видео по URL
	videoPath := filepath.Join("temp_videos", fmt.Sprintf("segment_%d.%s", segmentIndex, vg.Config.AI.Video.OutputFormat))
	err = downloadFile(responseData.VideoURL, videoPath, vg.Logger)
	if err != nil {
		return "", fmt.Errorf("ошибка при скачивании видео: %w", err)
	}

	vg.Logger.Info("Видеофрагмент для сцены %d сгенерирован и сохранен: %s", segmentIndex, videoPath)
	return videoPath, nil
}

// downloadFile скачивает файл с заданного URL и сохраняет его по указанному пути.
func downloadFile(url, filepath string, logger *utils.Logger) error {
	logger.Info("Скачивание файла: %s в %s", url, filepath)

	err := os.MkdirAll(filepath[:strings.LastIndex(filepath, "/")], os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("не удалось создать директорию: %w", err)
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("ошибка при HTTP GET запросе для скачивания: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("получен некорректный статус при скачивании: %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла для сохранения видео: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка записи скачанного файла: %w", err)
	}
	return nil
}
