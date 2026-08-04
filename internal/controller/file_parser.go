package controller

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robinskaba/rbxpmux/internal/config"
)

func parseTXTFile(path string) (string, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	var copy strings.Builder
	var remove strings.Builder
	buf := &copy

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		switch strings.ToLower(line) {
		case "copy":
			buf = &copy
		case "remove":
			buf = &remove
		default:
			commentIdx := strings.Index(line, "--")
			if commentIdx != -1 {
				line = line[:commentIdx]
			}
			line = strings.TrimRight(line, " ")
			buf.WriteString(line + "\n")
		}
	}
	return strings.TrimSpace(copy.String()), strings.TrimSpace(remove.String()), nil
}

func archiveTXTFile(universeId int, copyStr, removeStr string) error {
	dir, err := config.GetProgramDir()
	if err != nil {
		return err
	}

	historyDir := filepath.Join(dir, "history", fmt.Sprint(universeId))
	err = os.MkdirAll(historyDir, 0755)
	if err != nil {
		return err
	}

	now := time.Now().Format("02-01-2006_15-04")
	destPath := filepath.Join(historyDir, now+".txt")

	content := copyStr
	if removeStr != "" {
		if content != "" {
			content += "\n\n"
		}
		content += removeStr
	}

	err = os.WriteFile(destPath, []byte(content), 0644)
	if err == nil {
		slog.Info("instruction file archived", "path", destPath)
	}
	return err
}
