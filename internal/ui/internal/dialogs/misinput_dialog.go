package dialogs

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
)

func BadInput(fieldName string, parent fyne.Window) {
	dialog.ShowCustom(
		fmt.Sprintf("Invalid %s", fieldName),
		"Okay",
		canvas.NewText(
			fmt.Sprintf("Please enter a valid %s.", fieldName),
			theme.Color(theme.ColorNamePrimary),
		),
		parent,
	)
}
