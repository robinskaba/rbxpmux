package ui

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/robinskaba/rbxpmux/internal/controller"
	"github.com/robinskaba/rbxpmux/internal/editor"
	"github.com/robinskaba/rbxpmux/internal/ui/internal/views"
)

func startPublishProcess(controller *controller.Controller, universeId int, origin views.Place, targets []views.Place, copyStr string, removeStr string) {
	if controller.Settings.ApiKey == "" {
		dialog.ShowInformation("Missing API key", "You do not have an API key configured.", mainWindow)
		return
	}

	vbox := container.NewVBox()

	type stepUI struct {
		icon  *widget.Icon
		label *widget.Label
	}

	addStep := func(text string) *stepUI {
		icon := widget.NewIcon(theme.RadioButtonIcon())
		label := widget.NewLabel(text)
		step := &stepUI{icon: icon, label: label}
		vbox.Add(container.NewHBox(icon, label))
		return step
	}

	vbox.Add(widget.NewLabelWithStyle("Downloading", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	originDownloadStep := addStep(origin.Name)

	targetDownloadSteps := make(map[int]*stepUI)
	for _, target := range targets {
		targetDownloadSteps[target.Id] = addStep(target.Name)
	}

	vbox.Add(widget.NewSeparator())
	vbox.Add(widget.NewLabelWithStyle("Modifying", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))

	targetModifySteps := make(map[int]*stepUI)
	for _, target := range targets {
		targetModifySteps[target.Id] = addStep(target.Name)
	}

	vbox.Add(widget.NewSeparator())
	vbox.Add(widget.NewLabelWithStyle("Publishing", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))

	targetPublishSteps := make(map[int]*stepUI)
	for _, target := range targets {
		targetPublishSteps[target.Id] = addStep(target.Name)
	}

	scroll := container.NewVScroll(vbox)
	scroll.SetMinSize(fyne.NewSize(500, 450))

	closeBtn := widget.NewButton("Close", nil)
	closeBtn.Hide()

	content := container.NewBorder(nil, closeBtn, nil, nil, scroll)
	dlg := dialog.NewCustomWithoutButtons("Publishing Places", content, mainWindow)

	closeBtn.OnTapped = func() {
		dlg.Hide()
	}

	dlg.Show()

	go func() {
		markDone := func(step *stepUI) {
			fyne.Do(func() {
				step.icon.SetResource(theme.NewColoredResource(theme.ConfirmIcon(), theme.ColorNameSuccess))
				step.icon.Refresh()
			})
		}

		markError := func(step *stepUI, err error) {
			fyne.Do(func() {
				step.icon.SetResource(theme.NewColoredResource(theme.CancelIcon(), theme.ColorNameError))
				step.icon.Refresh()
				step.label.Text += fmt.Sprintf(" - %s", err.Error())
				step.label.Refresh()
			})
		}

		var wg sync.WaitGroup
		var processErr error
		var errMu sync.Mutex

		setError := func(err error) {
			errMu.Lock()
			if processErr == nil {
				processErr = err
			}
			errMu.Unlock()
		}

		wg.Go(func() {
			err := controller.DownloadOrigin(origin.Id)
			if err != nil {
				setError(err)
				markError(originDownloadStep, err)
				return
			}
			markDone(originDownloadStep)
		})

		for _, target := range targets {
			wg.Add(1)
			go func(tId int) {
				defer wg.Done()
				err := controller.DownloadTarget(tId)
				if err != nil {
					setError(err)
					markError(targetDownloadSteps[tId], err)
					return
				}
				markDone(targetDownloadSteps[tId])
			}(target.Id)
		}
		wg.Wait()

		if processErr != nil {
			slog.Error("process failed with error", "error", processErr)
			fyne.Do(func() {
				dialog.ShowError(processErr, mainWindow)
				closeBtn.Show()
			})
			return
		}

		err := controller.EditPlaces(copyStr, removeStr)
		if err != nil {
			isExpected := errors.Is(err, editor.ErrInstanceNotFound) ||
				errors.Is(err, editor.ErrMissingInOrigin) ||
				errors.Is(err, editor.ErrServiceRemoval) ||
				errors.Is(err, editor.ErrMultipleMatches)

			if !isExpected {
				slog.Error("unexpected error modifying places", "error", err)
				fyne.Do(func() {
					dialog.ShowError(err, mainWindow)
				})
			}

			for _, step := range targetModifySteps {
				markError(step, err)
			}
			closeBtn.Show()
			return
		}

		for _, step := range targetModifySteps {
			markDone(step)
		}

		for _, target := range targets {
			err := controller.PublishPlace(universeId, target.Id)
			if err != nil {
				setError(err)
				markError(targetPublishSteps[target.Id], err)
				continue
			}
			markDone(targetPublishSteps[target.Id])
		}

		wg.Wait()

		if processErr != nil {
			slog.Error("publish failed with error", "error", processErr)
			fyne.Do(func() {
				dialog.ShowError(processErr, mainWindow)
			})
		} else {
			err := controller.ArchiveTXTFile(universeId, copyStr, removeStr)
			if err != nil {
				slog.Warn("failed to archive instruction file", "error", err)
			}
			fyne.Do(func() {
				closeBtn.SetText("Published successfully")
				closeBtn.SetIcon(theme.ConfirmIcon())
				closeBtn.Importance = widget.SuccessImportance
				closeBtn.Refresh()
			})
		}

		fyne.Do(func() {
			closeBtn.Show()
		})
	}()
}
