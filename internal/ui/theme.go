package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type customTheme struct {
	fyne.Theme
}

func (m *customTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNamePrimary {
		if variant == theme.VariantLight {
			return color.Black
		}
		return color.White
	}
	if name == theme.ColorNameForegroundOnPrimary {
		if variant == theme.VariantLight {
			return color.White
		}
		return color.Black
	}
	return m.Theme.Color(name, variant)
}
