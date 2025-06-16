// internal/config/config.go
package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// AppConfig содержит общие настройки приложения, загружаемые из YAML.
type AppConfig struct {
	AI struct {
		Text struct {
			Model             string  `yaml:"model"`
			MaxTokensGeneral  int     `yaml:"max_tokens_general"`
			MaxTokensDetailed int     `yaml:"max_tokens_detailed"`
			Temperature       float64 `yaml:"temperature"`
		} `yaml:"text"`
		Video struct {
			OutputFormat string `yaml:"output_format"`
			Resolution   string `yaml:"resolution"`
			FPS          int    `yaml:"fps"`
		} `yaml:"video"`
	} `yaml:"ai"`
	// Здесь больше нет секции Platforms, так как ключи будут в .env
}

// Config содержит все настройки для нашего бота, включая переменные среды и YAML.
type Config struct {
	AppName         string
	YouTubeAPIKey   string
	TikTokAPIKey    string // Теперь ключ TikTok тоже здесь, из .env
	TextAIEndpoint  string
	VideoAIEndpoint string
	VideoAIAPIKey   string
	App             *AppConfig // Ссылка на YAML-конфигурацию
}

// LoadConfig загружает конфигурацию из переменных среды и YAML файла.
func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Предупреждение: файл .env не найден. Загрузка переменных среды напрямую.")
	}

	// Загрузка YAML-конфигурации
	yamlFile, err := ioutil.ReadFile("config/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения config.yaml: %w", err)
	}

	var appCfg AppConfig
	err = yaml.Unmarshal(yamlFile, &appCfg)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга config.yaml: %w", err)
	}

	cfg := &Config{
		AppName:         getEnv("APP_NAME", "YouTube Shorts AI Bot"),
		YouTubeAPIKey:   os.Getenv("YOUTUBE_API_KEY"),
		TikTokAPIKey:    os.Getenv("TIKTOK_API_KEY"), // Читаем ключ TikTok из .env
		TextAIEndpoint:  getEnv("TEXT_AI_ENDPOINT", "http://10.66.66.5:8000/v1/chat/completions"),
		VideoAIEndpoint: getEnv("VIDEO_AI_ENDPOINT", "http://10.66.66.5:8081/v1/video/generations"),
		VideoAIAPIKey:   os.Getenv("VIDEO_AI_API_KEY"),
		App:             &appCfg, // Сохраняем загруженную YAML-конфигурацию
	}

	// Базовые проверки, что ключи API и эндпоинты не пустые
	if cfg.YouTubeAPIKey == "" {
		fmt.Println("Предупреждение: YOUTUBE_API_KEY не установлен. Загрузка на YouTube может быть невозможна.")
	}
	if cfg.TikTokAPIKey == "" { // Проверяем ключ TikTok из .env
		fmt.Println("Предупреждение: TIKTOK_API_KEY не установлен. Загрузка на TikTok может быть невозможна.")
	}
	if cfg.TextAIEndpoint == "" {
		return nil, fmt.Errorf("TEXT_AI_ENDPOINT не установлен")
	}
	if cfg.VideoAIEndpoint == "" {
		return nil, fmt.Errorf("VIDEO_AI_ENDPOINT не установлен")
	}
	if cfg.VideoAIAPIKey == "" {
		return nil, fmt.Errorf("VIDEO_AI_API_KEY не установлен")
	}

	return cfg, nil
}

// getEnv получает переменную среды или возвращает значение по умолчанию.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
