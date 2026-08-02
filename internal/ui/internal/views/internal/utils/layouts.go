package utils

import "fyne.io/fyne/v2"

// centered with normal size

type CenterHalfLayout struct{}

func (c *CenterHalfLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}

	targetWidth := size.Width * 0.5
	targetHeight := objects[0].MinSize().Height

	x := (size.Width - targetWidth) / 2
	y := (size.Height - targetHeight) / 2

	objects[0].Resize(fyne.NewSize(targetWidth, targetHeight))
	objects[0].Move(fyne.NewPos(x, y))
}

func (c *CenterHalfLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}
	return objects[0].MinSize()
}
