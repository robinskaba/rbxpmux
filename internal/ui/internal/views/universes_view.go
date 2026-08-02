package views

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/robinskaba/rbxpmux/internal/ui/internal/views/internal/utils"
)

type UniverseConfiguration struct {
	Id           int
	Name         string
	ThumbnailURL string
}

type UniverseViewConfiguration struct {
	Universes              []UniverseConfiguration
	OnUniverseSelected     func(int)
	OnPrivateUniverseInput func(string)
}

func UniversesView(cfg *UniverseViewConfiguration) fyne.CanvasObject {
	heading := canvas.NewText("Select a universe", theme.Color(theme.ColorNamePrimary))
	heading.TextSize = 28
	heading.TextStyle.Bold = true
	heading.Alignment = fyne.TextAlignCenter
	elseLabel := utils.Label("or enter a private universe ID")
	elseLabel.Alignment = fyne.TextAlignCenter
	elseLabel.Color = theme.Color(theme.ColorNamePrimary)

	var universeCards []fyne.CanvasObject
	for _, universe := range cfg.Universes {
		universeCards = append(universeCards, customCard(
			universe.ThumbnailURL,
			universe.Name,
			func() { cfg.OnUniverseSelected(universe.Id) },
		))
	}

	elseInput := widget.NewEntry()
	enter := widget.NewButton("Confirm", func() {
		cfg.OnPrivateUniverseInput(elseInput.Text)
	})

	gap := canvas.NewRectangle(color.Transparent)
	gap.SetMinSize(fyne.NewSize(0, 10))

	return container.New(&utils.CenterHalfLayout{}, container.NewVBox(
		heading,
		gap,
		universeList(universeCards),
		gap,
		elseLabel,
		elseInput,
		enter,
	))
}

func customCard(imgURL string, name string, onTapped func()) fyne.CanvasObject {
	img := canvas.NewImageFromResource(utils.SafeURLResource(imgURL))
	img.FillMode = canvas.ImageFillStretch

	lbl := widget.NewLabel(name)
	lbl.Alignment = fyne.TextAlignCenter
	lbl.Wrapping = fyne.TextWrapWord

	textMinHeight := canvas.NewRectangle(color.Transparent)
	textMinHeight.SetMinSize(fyne.NewSize(0, 65))
	textContainer := container.NewStack(textMinHeight, lbl)

	content := container.NewBorder(nil, textContainer, nil, nil, img)

	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bg.CornerRadius = theme.InputRadiusSize()

	stack := container.NewStack(bg, content)

	c := &utils.ClickableCard{
		Content:  stack,
		OnTapped: onTapped,
	}
	c.ExtendBaseWidget(c)

	return container.NewCenter(c)
}

func universeList(cards []fyne.CanvasObject) fyne.CanvasObject {
	return container.NewHScroll(container.NewHBox(cards...))
}
