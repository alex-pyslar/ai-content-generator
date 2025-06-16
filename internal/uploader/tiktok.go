// internal/uploader/tiktok.go
package uploader

import (
	"fmt"
	"path/filepath"

	"ai-content-gen/pkg/utils"
)

// TikTokUploader implements VideoUploader for TikTok.
type TikTokUploader struct {
	APIKey string // TikTok APIKey или другой токен
	Logger *utils.Logger
}

// NewTikTokUploader creates a new TikTokUploader instance.
func NewTikTokUploader(apiKey string, logger *utils.Logger) *TikTokUploader {
	return &TikTokUploader{
		APIKey: apiKey,
		Logger: logger,
	}
}

// Upload загружает видеофайл на TikTok.
// В реальной реализации здесь будет использование TikTok for Developers API.
func (t *TikTokUploader) Upload(videoPath, title, description, tags string) (string, error) {
	t.Logger.Info("Начало загрузки видео на TikTok: %s", videoPath)
	t.Logger.Info("Название: %s, Описание: %s, Теги: %s", title, description, tags)

	if t.APIKey == "" {
		t.Logger.Warn("TikTok API Key не предоставлен. Загрузка на TikTok невозможна.")
		return "", fmt.Errorf("TikTok API Key не предоставлен")
	}

	// --- ЗАГЛУШКА ДЛЯ TikTok ---
	t.Logger.Warn("Внимание: Загрузка видео на TikTok - это заглушка. Реализуйте использование TikTok API здесь.")
	dummyVideoID := fmt.Sprintf("dummy_tiktok_id_%s", filepath.Base(videoPath))
	return fmt.Sprintf("https://www.tiktok.com/@youraccount/video/%s", dummyVideoID), nil
	// --- КОНЕЦ ЗАГЛУШКИ ---
}
