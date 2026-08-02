package views

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/robinskaba/rbxpmux/internal/ui/internal/views/internal/utils"
)

var requiredScopes = []string{
	"legacy-asset:manage",
	"universe-places:write",
	"universe.place:read",
}

type ApiConfigurationViewConfiguration struct {
	OnSubmitCallback func(userId string, apiKey string)
}

func ApiConfigurationView(cfg *ApiConfigurationViewConfiguration) fyne.CanvasObject {
	heading := canvas.NewText("Authorization required", theme.Color(theme.ColorNamePrimary))
	heading.TextSize = 24
	heading.TextStyle.Bold = true
	heading.Alignment = fyne.TextAlignCenter

	userIdArea := widget.NewEntry()
	apiKeyArea := widget.NewMultiLineEntry()
	submit := widget.NewButton("Save", func() {
		cfg.OnSubmitCallback(userIdArea.Text, apiKeyArea.Text)
	})

	scopesText := canvas.NewText(strings.Join(requiredScopes, ", "), theme.Color(theme.ColorNamePrimary))
	scopesText.TextSize = 12
	scopesText.TextStyle.Italic = true

	configurationForm := container.NewVBox(
		heading,
		container.NewPadded(utils.Label("User ID")),
		userIdArea,
		container.NewPadded(utils.Label("API Key")),
		container.NewPadded(scopesText),
		apiKeyArea,
		submit,
	)

	view := container.New(&utils.CenterHalfLayout{}, configurationForm)

	return view
}
