// internal/video/editor.go
package video

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"ai-content-gen/pkg/utils"
)

// VideoEditor отвечает за обработку и склейку видео.
type VideoEditor struct {
	Logger *utils.Logger
}

// NewVideoEditor создает новый экземпляр VideoEditor.
func NewVideoEditor(logger *utils.Logger) *VideoEditor {
	return &VideoEditor{
		Logger: logger,
	}
}

// ConcatenateVideos склеивает список видеофайлов в один с помощью FFmpeg.
// inputPaths: список путей к видеофайлам для склейки.
// outputPath: путь, куда будет сохранен склеенный файл.
// fps: частота кадров для выходного видео (важно для Shorts).
func (ve *VideoEditor) ConcatenateVideos(inputPaths []string, outputPath string, fps int) (string, error) {
	ve.Logger.Info("Начало склейки видеофайлов с FFmpeg: %v в %s", inputPaths, outputPath)

	if len(inputPaths) == 0 {
		return "", fmt.Errorf("нет входных видеофайлов для склейки")
	}

	// Создаем временный файл-список для FFmpeg
	listFilePath := "temp_videos/concat_list.txt"
	listFile, err := os.Create(listFilePath)
	if err != nil {
		return "", fmt.Errorf("не удалось создать список файлов для FFmpeg: %w", err)
	}
	defer listFile.Close()
	defer os.Remove(listFilePath) // Удаляем временный файл после использования

	for _, path := range inputPaths {
		_, err := listFile.WriteString(fmt.Sprintf("file '%s'\n", filepath.ToSlash(path))) // FFmpeg предпочитает /
		if err != nil {
			return "", fmt.Errorf("не удалось записать в список файлов для FFmpeg: %w", err)
		}
	}
	listFile.Close() // Закрыть, чтобы FFmpeg мог его прочитать

	// Создаем директорию для выходного видео, если ее нет
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil && !os.IsExist(err) {
		return "", fmt.Errorf("не удалось создать выходную директорию %s: %w", outputDir, err)
	}

	// Команда FFmpeg для склейки
	// -f concat: указывает формат входного файла как "concat" (для списка файлов)
	// -safe 0: разрешает произвольные пути в файле списка (может быть небезопасно, но нужно для некоторых путей)
	// -i listFilePath: указывает файл-список как входной
	// -c copy: копирует потоки без перекодирования (быстро, но требует, чтобы все входные видео были одинаковы)
	// Если видео разные, нужно перекодировать: -c:v libx264 -preset fast -crf 23 -c:a aac -b:a 128k
	// Для Shorts часто важна ориентация и FPS, возможно, придется добавить фильтры:
	// -vf "fps=30,scale=1080:1920:force_original_aspect_ratio=increase,crop=1080:1920"
	// (Для Shorts: 1080x1920, 30 FPS, вертикальное видео)
	// Я добавляю FPS, остальные параметры можно добавить в config.yaml и передавать.
	cmdArgs := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", listFilePath,
		"-c", "copy",
		"-r", fmt.Sprintf("%d", fps), // Устанавливаем FPS
		outputPath,
	}

	ve.Logger.Info("Запуск FFmpeg с командой: ffmpeg %s", strings.Join(cmdArgs, " "))
	cmd := exec.Command("ffmpeg", cmdArgs...)

	// Захват вывода FFmpeg для отладки
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		ve.Logger.Error("Ошибка FFmpeg. Stdout: %s", stdout.String())
		ve.Logger.Error("Ошибка FFmpeg. Stderr: %s", stderr.String())
		return "", fmt.Errorf("ошибка выполнения команды FFmpeg: %w", err)
	}

	ve.Logger.Info("FFmpeg успешно завершил склейку. Stdout: %s", stdout.String())
	ve.Logger.Info("Видео успешно склеено в: %s", outputPath)
	return outputPath, nil
}
