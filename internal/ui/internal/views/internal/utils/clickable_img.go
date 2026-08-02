package utils

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type clickableImage struct {
	widget.BaseWidget
	img      *canvas.Image
	OnTapped func()
}

func (c *clickableImage) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.img)
}

func (c *clickableImage) MinSize() fyne.Size {
	return widget.NewButtonWithIcon("", theme.SettingsIcon(), nil).MinSize() // steal dimensions from a regular button
}

func (c *clickableImage) Tapped(_ *fyne.PointEvent) {
	if c.OnTapped != nil {
		c.OnTapped()
	}
}

func (c *clickableImage) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func ButtonWithFullImage(res fyne.Resource, onClick func()) fyne.CanvasObject {
	img := canvas.NewImageFromResource(res)
	img.FillMode = canvas.ImageFillStretch
	c := &clickableImage{
		img:      img,
		OnTapped: onClick,
	}
	c.ExtendBaseWidget(c)

	return c
}
