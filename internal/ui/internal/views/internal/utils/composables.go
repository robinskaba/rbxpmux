package utils

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// text

func Label(text string) *canvas.Text {
	label := canvas.NewText(text, theme.Color(theme.ColorNamePrimary))
	label.TextSize = 16
	return label
}

// cards

type ClickableCard struct {
	widget.BaseWidget
	Content  fyne.CanvasObject
	OnTapped func()
}

func (c *ClickableCard) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.Content)
}

func (c *ClickableCard) Tapped(_ *fyne.PointEvent) {
	if c.OnTapped != nil {
		c.OnTapped()
	}
}

func (c *ClickableCard) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func (c *ClickableCard) MinSize() fyne.Size {
	return fyne.NewSize(200, 260)
}
