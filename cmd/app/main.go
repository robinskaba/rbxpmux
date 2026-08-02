package main

import (
	_ "embed"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/robinskaba/rbxpmux/internal/client"
	"github.com/robinskaba/rbxpmux/internal/config"
	"github.com/robinskaba/rbxpmux/internal/controller"
	"github.com/robinskaba/rbxpmux/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

//go:embed Icon.png
var iconData []byte

func main() {
	logDir, _ := config.GetProgramDir()
	logPath := filepath.Join(logDir, "logs")
	os.MkdirAll(logPath, 0755)
	now := time.Now().Format("02-01-2006_15-04-05")
	logFile, err := os.OpenFile(filepath.Join(logPath, "rbxpmux_"+now+".log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		slog.SetDefault(slog.New(slog.NewTextHandler(logFile, nil)))
	}

	settings, err := config.LoadSettings()
	if err != nil {
		slog.Error("failed to load settings", "error", err)
		os.Exit((1))
	}
	configuration, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	api := client.NewRbxApi(settings.ApiKey)
	controller := controller.NewController(api, settings, configuration)

	a := app.NewWithID("rbxpmux")

	icon := fyne.NewStaticResource("Icon.png", iconData)
	a.SetIcon(icon)
	w := ui.MainWindow(a, controller)
	w.Show()
	a.Run()
}
