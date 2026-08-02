package config

import (
	"os"
	"path/filepath"
)

func GetProgramDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appDir := filepath.Join(dir, "rbxpmux")
	err = os.MkdirAll(appDir, 0755) // read+write+run permissions
	if err != nil {
		return "", err
	}

	return appDir, nil
}
