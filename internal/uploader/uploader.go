// internal/uploader/uploader.go
package uploader

import (
	"fmt"

	"ai-content-gen/pkg/utils"
)

// PlatformType определяет тип платформы для загрузки.
type PlatformType string

const (
	PlatformYouTube PlatformType = "youtube"
	PlatformTikTok  PlatformType = "tiktok"
)

// VideoUploader определяет интерфейс для загрузки видео на конкретную платформу.
type VideoUploader interface {
	Upload(videoPath, title, description, tags string) (string, error)
}

// MultiPlatformUploader управляет загрузкой на различные платформы.
type MultiPlatformUploader struct {
	platforms map[PlatformType]VideoUploader
	Logger    *utils.Logger
}

// NewMultiPlatformUploader создает новый экземпляр MultiPlatformUploader.
// Принимает API-ключи напрямую, так как они будут загружены в main из .env.
func NewMultiPlatformUploader(youtubeAPIKey, tiktokAPIKey string, logger *utils.Logger) *MultiPlatformUploader {
	m := &MultiPlatformUploader{
		platforms: make(map[PlatformType]VideoUploader),
		Logger:    logger,
	}

	// Инициализируем загрузчики для каждой платформы
	m.platforms[PlatformYouTube] = NewYouTubeUploader(youtubeAPIKey, logger)
	m.platforms[PlatformTikTok] = NewTikTokUploader(tiktokAPIKey, logger)

	return m
}

// Upload загружает видео на указанную платформу.
func (m *MultiPlatformUploader) Upload(platform PlatformType, videoPath, title, description, tags string) (string, error) {
	uploader, ok := m.platforms[platform]
	if !ok {
		return "", fmt.Errorf("загрузчик для платформы %s не найден", platform)
	}

	m.Logger.Info("Запуск загрузки видео '%s' на платформу: %s", videoPath, platform)
	url, err := uploader.Upload(videoPath, title, description, tags)
	if err != nil {
		m.Logger.Error("Ошибка загрузки на %s: %v", platform, err)
		return "", err
	}
	m.Logger.Info("Видео успешно загружено на %s. URL: %s", platform, url)
	return url, nil
}
