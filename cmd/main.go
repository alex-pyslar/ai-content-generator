// cmd/main.go
package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"ai-content-gen/internal/ai"
	"ai-content-gen/internal/config"
	"ai-content-gen/internal/uploader"
	"ai-content-gen/internal/video"
	"ai-content-gen/pkg/utils"
)

func main() {
	// Инициализируем логгер первым делом
	logger := utils.NewLogger()
	logger.Info("Запуск YouTube Shorts AI Bot...")

	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Ошибка загрузки конфигурации: %v", err)
	}

	logger.Info("Бот запущен с настройкой: %s", cfg.AppName)
	logger.Info("Эндпоинт текстового ИИ: %s", cfg.TextAIEndpoint)
	logger.Info("Эндпоинт видео ИИ: %s", cfg.VideoAIEndpoint)
	logger.Info("Модель текстового ИИ: %s", cfg.App.AI.Text.Model)

	// Инициализация сервисов
	textGen := ai.NewTextGenerator(cfg.TextAIEndpoint, cfg.App, logger)
	videoGen := ai.NewVideoGenerator(cfg.VideoAIEndpoint, cfg.VideoAIAPIKey, cfg.App, logger)
	videoEditor := video.NewVideoEditor(logger)

	// Инициализация мультиплатформенного загрузчика с ключами из конфига
	multiUploader := uploader.NewMultiPlatformUploader(
		cfg.YouTubeAPIKey,
		cfg.TikTokAPIKey,
		logger,
	)

	topic := "космическая битва с флотом Федерации"

	// 1. Генерируем общую идею и краткое описание сцен
	generalContent, err := textGen.GenerateShortsIdeaAndScenes(topic)
	if err != nil {
		logger.Fatal("Ошибка при генерации общей идеи и сцен: %v", err)
	}

	logger.Info("\n--- Сгенерированная общая идея и сцены ---")
	fmt.Println(generalContent)
	logger.Info("----------------------------------------")

	overallIdea, sceneDescriptions := parseIdeaAndScenes(generalContent, logger)
	if overallIdea == "" || len(sceneDescriptions) == 0 {
		logger.Fatal("Не удалось извлечь идею или описания сцен из сгенерированного контента.")
	}

	logger.Info("Общая идея: %s", overallIdea)
	for i, scene := range sceneDescriptions {
		logger.Info("Сцена %d: %s", i+1, scene)
	}

	// Создаем директорию для временных видео
	if err := os.MkdirAll("temp_videos", os.ModePerm); err != nil {
		logger.Fatal("Не удалось создать директорию temp_videos: %v", err)
	}
	defer func() { // Добавляем defer для очистки временных файлов
		logger.Info("Очистка временных видеофайлов...")
		if err := os.RemoveAll("temp_videos"); err != nil {
			logger.Warn("Не удалось удалить временную папку 'temp_videos': %v", err)
		}
		logger.Info("Временные видеофайлы удалены.")
	}()

	// 2. Для каждой сцены генерируем подробный промпт для видео
	logger.Info("\n--- Генерация подробных промптов для видео ---")
	var detailedPrompts []string
	for i, sceneDesc := range sceneDescriptions {
		detailedPrompt, err := textGen.GenerateVideoPromptForScene(overallIdea, sceneDesc)
		if err != nil {
			logger.Error("Ошибка при генерации подробного промпта для сцены %d: %v", i+1, err)
			continue
		}
		detailedPrompts = append(detailedPrompts, detailedPrompt)
		logger.Info("Детальный промпт для Сцены %d:\n%s\n", i+1, detailedPrompt)
		logger.Info("-------------------------------------------")
	}

	if len(detailedPrompts) == 0 {
		logger.Fatal("Не удалось сгенерировать ни одного детального промпта.")
	}

	// 3. Генерируем все видеосегменты на основе детальных промптов
	logger.Info("\n--- Генерация всех видеосегментов ---")
	var videoSegmentPaths []string
	for i, prompt := range detailedPrompts {
		segmentPath, err := videoGen.GenerateVideoSegment(prompt, i+1)
		if err != nil {
			logger.Error("Ошибка при генерации видео для сцены %d: %v", i+1, err)
			continue
		}
		videoSegmentPaths = append(videoSegmentPaths, segmentPath)
		logger.Info("Видеофрагмент для Сцены %d сгенерирован и сохранен: %s", i+1, segmentPath)
		logger.Info("-------------------------------------------")
	}

	if len(videoSegmentPaths) == 0 {
		logger.Fatal("Не удалось сгенерировать ни одного видеофрагмента.")
	}

	// 4. Склеиваем видеосегменты в одно финальное видео
	outputDir := "output_shorts"
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		logger.Fatal("Не удалось создать директорию %s: %v", outputDir, err)
	}
	finalVideoPath := fmt.Sprintf("%s/%s_final_short.%s", outputDir, strings.ReplaceAll(overallIdea, " ", "_"), cfg.App.AI.Video.OutputFormat)

	compiledVideoPath, err := videoEditor.ConcatenateVideos(videoSegmentPaths, finalVideoPath, cfg.App.AI.Video.FPS)
	if err != nil {
		logger.Fatal("Ошибка при склейке видео: %v", err)
	}
	logger.Info("Финальное видео скомпилировано: %s", compiledVideoPath)

	// Очистка временных видеофайлов после склейки
	logger.Info("Удаление отдельных видеосегментов после склейки...")
	for _, path := range videoSegmentPaths {
		if err := os.Remove(path); err != nil {
			logger.Warn("Не удалось удалить временный файл %s: %v", path, err)
		}
	}
	logger.Info("Отдельные видеосегменты удалены.")

	// 5. Отправляем это ОДНО финальное видео на ВСЕ нужные платформы
	logger.Info("\n--- Загрузка финального видео на платформы ---")

	// Метаданные для YouTube
	youtubeTitle := fmt.Sprintf("AI Shorts: %s", overallIdea)
	youtubeDescription := fmt.Sprintf("Это YouTube Shorts, сгенерированный полностью AI на тему: %s.", overallIdea)
	youtubeTags := "AI,Shorts,YouTubeShorts,AIgenerated,космическаябитва,федерация"

	// Загрузка на YouTube
	ytVideoURL, err := multiUploader.Upload(uploader.PlatformYouTube, compiledVideoPath, youtubeTitle, youtubeDescription, youtubeTags)
	if err != nil {
		logger.Error("Ошибка при загрузке видео на YouTube: %v", err) // Не фатально, если хотим попробовать другие платформы
	} else {
		logger.Info("Видео успешно загружено на YouTube: %s", ytVideoURL)
	}

	// Метаданные для TikTok
	tiktokTitle := fmt.Sprintf("AI Космос: %s", overallIdea)
	tiktokDescription := "Генерация AI для TikTok! #AI #Shorts"
	tiktokTags := "AI,космос,shorts"

	// Загрузка на TikTok (пример)
	tiktokVideoURL, err := multiUploader.Upload(uploader.PlatformTikTok, compiledVideoPath, tiktokTitle, tiktokDescription, tiktokTags)
	if err != nil {
		logger.Error("Ошибка при загрузке видео на TikTok: %v", err)
	} else {
		logger.Info("Видео успешно загружено на TikTok: %s", tiktokVideoURL)
	}

	logger.Info("Бот завершил свою работу!")
}

// parseIdeaAndScenes разбирает сгенерированный текст на общую идею и отдельные описания сцен.
func parseIdeaAndScenes(content string, logger *utils.Logger) (string, []string) {
	var overallIdea string
	var scenes []string

	lines := strings.Split(content, "\n")
	sceneRegex := regexp.MustCompile(`^Сцена \d+: (.+)`)
	ideaRegex := regexp.MustCompile(`^Идея: (.+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "Идея:") {
			if matches := ideaRegex.FindStringSubmatch(line); len(matches) > 1 {
				overallIdea = matches[1]
				logger.Info("Извлечена идея: %s", overallIdea)
			}
		} else if matches := sceneRegex.FindStringSubmatch(line); len(matches) > 1 {
			// Добавляем проверку на наличие содержимого после "Сцена N: "
			if len(matches[1]) > 0 {
				scenes = append(scenes, matches[1])
				logger.Info("Извлечена сцена: %s", matches[1])
			} else {
				logger.Warn("Пустое описание для сцены в строке: %s", line)
			}
		}
	}

	if overallIdea == "" {
		logger.Warn("Не удалось найти 'Идея:' в сгенерированном контенте.")
	}
	if len(scenes) == 0 {
		logger.Warn("Не удалось найти ни одной 'Сцены' в сгенерированном контенте.")
	}

	return overallIdea, scenes
}
