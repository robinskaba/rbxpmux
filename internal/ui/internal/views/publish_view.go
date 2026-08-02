package views

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/robinskaba/rbxpmux/internal/ui/internal/views/internal/utils"
)

type proportionLayout struct {
	split float32
}

func (l *proportionLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	pad := theme.Padding()
	availWidth := size.Width - pad
	leftWidth := float32(availWidth) * l.split

	objects[0].Resize(fyne.NewSize(leftWidth, size.Height))
	objects[0].Move(fyne.NewPos(0, 0))

	objects[1].Resize(fyne.NewSize(size.Width-leftWidth-pad, size.Height))
	objects[1].Move(fyne.NewPos(leftWidth+pad, 0))
}

func (l *proportionLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 2 {
		return fyne.NewSize(0, 0)
	}
	s1 := objects[0].MinSize()
	s2 := objects[1].MinSize()
	return fyne.NewSize(s1.Width+s2.Width+theme.Padding(), fyne.Max(s1.Height, s2.Height))
}

type Place struct {
	Id   int
	Name string
}

type PublishViewConfiguration struct {
	Places                []Place
	DefaultOrigin         Place
	DefaultTargets        []Place
	ThumbnailURL          string
	ActiveInstructionFile string
	// ui callbacks
	SetInputText func(copy, remove string)
	SetThumbnail func(string)
	// navigation
	OnUniverseClick   func()
	OnAuthClick       func()
	OnHistoryClick    func()
	OnFileUploadClick func()
	OnSettingsClick   func()
	OnPublish         func(Place, []Place, string, string)
}

var selectedOrigin int
var selectedTargets []int

func PublishView(cfg *PublishViewConfiguration) fyne.CanvasObject {
	topBar := topBar(
		&cfg.SetThumbnail,
		cfg.ThumbnailURL,
		cfg.OnUniverseClick,
		cfg.OnAuthClick,
		cfg.OnHistoryClick,
		cfg.OnFileUploadClick,
		cfg.OnSettingsClick,
	)

	placesMenu := placesMenu(cfg.Places, cfg.DefaultOrigin, cfg.DefaultTargets)
	inputMenu := inputMenu(func(copy, remove string) {
		var o Place
		var t []Place
		for _, p := range cfg.Places {
			if p.Id == selectedOrigin {
				o = p
			}
			for _, st := range selectedTargets {
				if p.Id == st {
					t = append(t, p)
				}
			}
		}
		cfg.OnPublish(o, t, copy, remove)
	}, &cfg.SetInputText)

	body := container.New(
		&proportionLayout{split: 0.3},
		placesMenu,
		inputMenu,
	)

	return container.NewBorder(
		topBar,
		nil,
		nil,
		nil,
		body,
	)
}

func inputMenu(onPublish func(string, string), setInputText *func(string, string)) fyne.CanvasObject {
	copyLabel := utils.Label("Copy")
	copyEntry := widget.NewMultiLineEntry()
	copyCol := container.NewBorder(copyLabel, nil, nil, nil, copyEntry)

	removeLabel := utils.Label("Remove")
	removeEntry := widget.NewMultiLineEntry()
	removeCol := container.NewBorder(removeLabel, nil, nil, nil, removeEntry)

	if setInputText != nil {
		*setInputText = func(copy, remove string) {
			copyEntry.SetText(copy)
			removeEntry.SetText(remove)
		}
	}

	publishBtn := widget.NewButton("Publish", func() {
		onPublish(copyEntry.Text, removeEntry.Text)
	})

	return container.NewBorder(
		nil,
		publishBtn,
		nil,
		nil,
		container.NewGridWithColumns(2, copyCol, removeCol),
	)
}

func placesMenu(places []Place, defaultOrigin Place, defaultTargets []Place) fyne.CanvasObject {
	var options []string
	placeMap := make(map[string]int)

	for _, p := range places {
		options = append(options, p.Name)
		placeMap[p.Name] = p.Id
	}

	originLabel := utils.Label("Origin place")
	singleSelect := widget.NewSelect(options, func(value string) {
		selectedOrigin = placeMap[value]
	})
	singleSelect.SetSelected(defaultOrigin.Name)
	targetLabel := utils.Label("Target places")
	multiSelect := widget.NewCheckGroup(options, func(values []string) {
		selectedTargets = []int{}
		for _, v := range values {
			selectedTargets = append(selectedTargets, placeMap[v])
		}
	})

	defaultTargetNames := make([]string, len(defaultTargets))
	for t := range defaultTargets {
		defaultTargetNames[t] = defaultTargets[t].Name
	}
	multiSelect.SetSelected(defaultTargetNames)

	scroll := container.NewVScroll(multiSelect)

	bg := canvas.NewRectangle(theme.InputBackgroundColor())
	bg.CornerRadius = theme.InputRadiusSize()
	bg.StrokeColor = theme.DisabledColor()
	bg.StrokeWidth = 1

	multiSelectWidget := container.NewStack(bg, scroll)

	topSection := container.NewVBox(
		originLabel,
		singleSelect,
		targetLabel,
	)

	content := container.NewBorder(
		topSection,
		nil,
		nil,
		nil,
		multiSelectWidget,
	)

	return content
}

func topBar(
	setThumbnail *func(string),
	initialThumbnailURL string,
	onUniverseClick func(),
	onAuthClick func(),
	onHistoryClick func(),
	onFileUploadClick func(),
	onSettingsClick func(),
) fyne.CanvasObject {
	// left side
	thumbnailWrapper := container.NewStack()

	updateThumbnail := func(url string) {
		newBtn := utils.ButtonWithFullImage(utils.SafeURLResource(url), onUniverseClick)
		thumbnailWrapper.Objects = []fyne.CanvasObject{newBtn}
		thumbnailWrapper.Refresh()
	}

	updateThumbnail(initialThumbnailURL)

	if setThumbnail != nil {
		*setThumbnail = updateThumbnail
	}

	// right side
	authCfgBtn := widget.NewButtonWithIcon("", theme.AccountIcon(), onAuthClick)
	historyBtn := widget.NewButtonWithIcon("", theme.HistoryIcon(), onHistoryClick)
	fileUploadBtn := widget.NewButtonWithIcon("", theme.FileTextIcon(), onFileUploadClick)
	settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), onSettingsClick)

	return container.NewHBox(
		thumbnailWrapper,
		layout.NewSpacer(),
		container.NewHBox(
			fileUploadBtn,
			historyBtn,
			authCfgBtn,
			settingsBtn,
		),
	)
}
