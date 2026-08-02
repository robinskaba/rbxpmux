package views

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/robinskaba/rbxpmux/internal/ui/internal/views/internal/utils"
)

type UniverseConfigViewConfiguration struct {
	Places         []Place
	DefaultOrigin  Place
	DefaultTargets []Place
	OnSave         func(originId int, targetIds []int)
	OnCancel       func()
}

func UniverseConfigView(cfg *UniverseConfigViewConfiguration) fyne.CanvasObject {
	var options []string
	placeMap := make(map[string]int)

	for _, p := range cfg.Places {
		options = append(options, p.Name)
		placeMap[p.Name] = p.Id
	}

	var selectedOrigin int
	if cfg.DefaultOrigin.Name != "" {
		selectedOrigin = cfg.DefaultOrigin.Id
	}

	var selectedTargets []int
	for _, t := range cfg.DefaultTargets {
		selectedTargets = append(selectedTargets, t.Id)
	}

	originLabel := utils.Label("Default origin")
	originSelect := widget.NewSelect(options, func(value string) {
		selectedOrigin = placeMap[value]
	})
	if cfg.DefaultOrigin.Name != "" {
		originSelect.SetSelected(cfg.DefaultOrigin.Name)
	}

	targetLabel := utils.Label("Default target places")
	targetsGroup := widget.NewCheckGroup(options, func(values []string) {
		selectedTargets = []int{}
		for _, v := range values {
			selectedTargets = append(selectedTargets, placeMap[v])
		}
	})

	var defaultTargetNames []string
	for _, t := range cfg.DefaultTargets {
		defaultTargetNames = append(defaultTargetNames, t.Name)
	}
	targetsGroup.SetSelected(defaultTargetNames)

	scroll := container.NewVScroll(targetsGroup)
	bg := canvas.NewRectangle(theme.InputBackgroundColor())
	bg.CornerRadius = theme.InputRadiusSize()
	bg.StrokeColor = theme.DisabledColor()
	bg.StrokeWidth = 1
	targetsWidget := container.NewStack(bg, scroll)

	saveBtn := widget.NewButton("Save", func() {
		if cfg.OnSave != nil {
			cfg.OnSave(selectedOrigin, selectedTargets)
		}
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		if cfg.OnCancel != nil {
			cfg.OnCancel()
		}
	})

	buttons := container.NewGridWithColumns(2, cancelBtn, saveBtn)

	topSection := container.NewVBox(
		originLabel,
		originSelect,
		targetLabel,
	)

	return container.NewBorder(
		topSection,
		buttons,
		nil,
		nil,
		targetsWidget,
	)
}
