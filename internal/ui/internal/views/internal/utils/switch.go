package utils

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Switch struct {
	widget.BaseWidget
	Checked   bool
	OnChanged func(bool)
}

func NewSwitch(onChanged func(bool)) *Switch {
	s := &Switch{OnChanged: onChanged}
	s.ExtendBaseWidget(s)
	return s
}

func (s *Switch) SetChecked(checked bool) {
	if s.Checked == checked {
		return
	}
	s.Checked = checked
	if s.OnChanged != nil {
		s.OnChanged(checked)
	}
	s.Refresh()
}

func (s *Switch) Tapped(_ *fyne.PointEvent) {
	s.SetChecked(!s.Checked)
}

func (s *Switch) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

type switchRenderer struct {
	s      *Switch
	bg     *canvas.Rectangle
	bubble *canvas.Circle
}

func (s *Switch) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(theme.DisabledColor())
	// Let Fyne handle the pill shape by setting radius to a large value, but
	// since CornerRadius is a float32 relative to dimensions, we'll set it dynamically in Layout
	bg.CornerRadius = 12

	bubble := canvas.NewCircle(theme.BackgroundColor())

	r := &switchRenderer{
		s:      s,
		bg:     bg,
		bubble: bubble,
	}
	r.Refresh()

	return r
}

func (r *switchRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.bg.Move(fyne.NewPos(0, 0))
	// make it fully rounded
	r.bg.CornerRadius = size.Height / 2

	bubbleSize := size.Height - 4
	r.bubble.Resize(fyne.NewSize(bubbleSize, bubbleSize))

	if r.s.Checked {
		r.bubble.Move(fyne.NewPos(size.Width-bubbleSize-2, 2))
	} else {
		r.bubble.Move(fyne.NewPos(2, 2))
	}
}

func (r *switchRenderer) MinSize() fyne.Size {
	return fyne.NewSize(36, 18)
}

func (r *switchRenderer) Refresh() {
	if r.s.Checked {
		r.bg.FillColor = theme.PrimaryColor()
	} else {
		r.bg.FillColor = theme.DisabledColor()
	}
	r.bg.Refresh()
	r.bubble.Refresh()
	r.Layout(r.s.Size())
}

func (r *switchRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.bubble}
}

func (r *switchRenderer) Destroy() {}

type buttonHeightLayout struct{}

func (l *buttonHeightLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) > 0 {
		objects[0].Resize(size)
		objects[0].Move(fyne.NewPos(0, 0))
	}
}

func (l *buttonHeightLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) > 0 {
		s := objects[0].MinSize()
		btnMin := widget.NewButton("T", nil).MinSize()
		return fyne.NewSize(s.Width, fyne.Max(s.Height, btnMin.Height))
	}
	return fyne.NewSize(0, 0)
}

func NewPublishSwitch(onChanged func(bool)) (*Switch, fyne.CanvasObject) {
	saveLbl := canvas.NewText("Save  ", theme.DisabledColor())
	saveLbl.TextSize = theme.TextSize()
	saveLbl.Alignment = fyne.TextAlignTrailing

	pubLbl := canvas.NewText("  Publish", theme.ForegroundColor())
	pubLbl.TextSize = theme.TextSize()

	updateLabels := func(checked bool) {
		if checked {
			saveLbl.Color = theme.DisabledColor()
			pubLbl.Color = theme.ForegroundColor()
		} else {
			saveLbl.Color = theme.ForegroundColor()
			pubLbl.Color = theme.DisabledColor()
		}
		saveLbl.Refresh()
		pubLbl.Refresh()
	}

	s := NewSwitch(func(checked bool) {
		updateLabels(checked)
		if onChanged != nil {
			onChanged(checked)
		}
	})
	
	hbox := container.NewHBox(saveLbl, container.NewCenter(s), pubLbl)
	cont := container.NewCenter(hbox)
	
	wrapper := container.New(&buttonHeightLayout{}, cont)
	return s, wrapper
}
