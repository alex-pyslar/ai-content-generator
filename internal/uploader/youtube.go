// internal/uploader/youtube.go
package uploader

import (
	"fmt"
	"path/filepath"

	"ai-content-gen/pkg/utils"
)

// YouTubeUploader implements VideoUploader for YouTube.
type YouTubeUploader struct {
	APIKey string
	Logger *utils.Logger
}

// NewYouTubeUploader creates a new YouTubeUploader instance.
func NewYouTubeUploader(apiKey string, logger *utils.Logger) *YouTubeUploader {
	return &YouTubeUploader{
		APIKey: apiKey,
		Logger: logger,
	}
}

// Upload загружает видеофайл на YouTube.
// В реальной реализации здесь будет использование YouTube Data API.
func (u *YouTubeUploader) Upload(videoPath, title, description, tags string) (string, error) {
	u.Logger.Info("Начало загрузки видео на YouTube: %s", videoPath)
	u.Logger.Info("Название: %s, Описание: %s, Теги: %s", title, description, tags)

	if u.APIKey == "" {
		u.Logger.Warn("YouTube API Key не предоставлен. Загрузка на YouTube невозможна.")
		return "", fmt.Errorf("YouTube API Key не предоставлен")
	}

	// --- ЗАГЛУШКА ДЛЯ YouTube ---
	u.Logger.Warn("Внимание: Загрузка видео на YouTube - это заглушка. Реализуйте использование YouTube Data API здесь.")
	dummyVideoID := fmt.Sprintf("dummy_youtube_id_%s", filepath.Base(videoPath))
	return fmt.Sprintf("http://youtube.com/watch?v=%s", dummyVideoID), nil
	// --- КОНЕЦ ЗАГЛУШКИ ---
}
