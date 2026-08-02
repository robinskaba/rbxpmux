package utils

import (
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func SafeURLResource(url string) fyne.Resource {
	res, err := fyne.LoadResourceFromURLString(url)
	if err != nil {
		if url != "" {
			slog.Warn("failed to load resource from URL", "url", url, "error", err)
		}
		return theme.BrokenImageIcon()
	} else {
		return res
	}
}
