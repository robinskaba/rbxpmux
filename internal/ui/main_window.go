package ui

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/robinskaba/rbxpmux/internal/client"
	"github.com/robinskaba/rbxpmux/internal/config"
	"github.com/robinskaba/rbxpmux/internal/controller"
	"github.com/robinskaba/rbxpmux/internal/ui/internal/dialogs"
	"github.com/robinskaba/rbxpmux/internal/ui/internal/views"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	sqdialog "github.com/sqweek/dialog"
)

var mainWindow fyne.Window

func MainWindow(a fyne.App, controller *controller.Controller) fyne.Window {
	a.Settings().SetTheme(&customTheme{Theme: theme.DefaultTheme()})
	mainWindow = a.NewWindow("Roblox Place Multiplexer")
	mainWindow.Resize(fyne.NewSize(1000, 600))

	if controller.Settings.UserId == 0 || controller.Settings.ApiKey == "" {
		setApiConfigurationView(controller)
	} else if controller.Configuration.LatestUniverseId == 0 {
		setUniverseSelectView(controller)
	} else {
		setPublishView(controller, controller.Configuration.LatestUniverseId, nil)
	}

	return mainWindow
}

func setApiConfigurationView(controller *controller.Controller) {
	apiConfigurationViewCfg := &views.ApiConfigurationViewConfiguration{
		OnSubmitCallback: func(userIdStr string, apiKey string) {
			userId, err := strconv.Atoi(userIdStr)
			if err != nil {
				dialogs.BadInput("User ID", mainWindow)
				return
			}

			controller.SaveApiConfig(userId, apiKey)
			setUniverseSelectView(controller)
		},
	}
	mainWindow.SetTitle("Authorization configuration - rbxpmux")
	mainWindow.SetContent(views.ApiConfigurationView(apiConfigurationViewCfg))
}

func setUniverseSelectView(controller *controller.Controller) {
	loadingBar := widget.NewProgressBarInfinite()
	loadingBar.Start()

	wrapper := container.NewStack(container.NewCenter(loadingBar))
	mainWindow.SetContent(wrapper)

	go func() {
		universes, err := controller.Api.GetUniverses(controller.Settings.UserId)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(err, mainWindow)
				loadingBar.Stop()
			})
			return
		}

		var configuration views.UniverseViewConfiguration
		for _, universe := range universes {
			thumbnail, err := controller.Api.GetUniverseThumbnail(universe.Id, client.MEDIUM_THUMBNAIL)
			if err != nil {
				slog.Warn("failed to retrieve universe thumbnail", "universe", universe.Name, "error", err)
			}
			configuration.Universes = append(configuration.Universes, views.UniverseConfiguration{
				Id:           universe.Id,
				Name:         universe.Name,
				ThumbnailURL: thumbnail,
			})
		}

		configuration.OnUniverseSelected = func(universeId int) {
			slog.Info("universe selected", "id", universeId)
			setPublishView(controller, universeId, nil)
		}
		configuration.OnPrivateUniverseInput = func(strId string) {
			id, err := strconv.Atoi(strId)
			if err != nil {
				dialogs.BadInput("Universe ID", mainWindow)
				return
			}
			configuration.OnUniverseSelected(id)
		}

		fyne.Do(func() {
			loadingBar.Stop()
			wrapper.Objects = []fyne.CanvasObject{views.UniversesView(&configuration)}
			mainWindow.SetTitle("Universe selection - rbxpmux")
			wrapper.Refresh()
		})
	}()
}

func setPublishView(controller *controller.Controller, universeId int, publishViewCfg *views.PublishViewConfiguration) {
	if publishViewCfg != nil {
		publishView := views.PublishView(publishViewCfg)
		mainWindow.SetContent(publishView)
		mainWindow.SetTitle(fmt.Sprintf("Place publishing in %d - rbxpmux", universeId))
		return
	}

	loadingBar := widget.NewProgressBarInfinite()
	loadingBar.Start()
	wrapper := container.NewStack(container.NewCenter(loadingBar))
	mainWindow.SetContent(wrapper)
	mainWindow.SetTitle(fmt.Sprintf("Loading places for universe %d - rbxpmux", universeId))

	go func() {
		places, err := controller.Api.GetPlaces(universeId)
		if err != nil || len(places) == 0 {
			fyne.Do(func() {
				errDlg := dialog.NewError(fmt.Errorf("Universe doesn't exist or has no places."), mainWindow)
				errDlg.SetOnClosed(func() {
					setUniverseSelectView(controller)
				})
				errDlg.Show()
			})
			return
		}

		clientPlaceToViewPlace := func(cp client.Place) views.Place {
			return views.Place{
				Id:   cp.Id,
				Name: cp.Name,
			}
		}
		viewPlaces := make([]views.Place, len(places))
		for i := range places {
			viewPlaces[i] = clientPlaceToViewPlace(places[i])
		}

		originId, targetsIds := controller.GetUniverseDefaults(universeId)
		var defaultOrigin views.Place
		var defaultTargets []views.Place

		for _, vp := range viewPlaces {
			if vp.Id == originId {
				defaultOrigin = vp
			}
			if slices.Contains(targetsIds, vp.Id) {
				defaultTargets = append(defaultTargets, vp)
			}
		}

		publishViewCfg = &views.PublishViewConfiguration{
			Places:          viewPlaces,
			DefaultOrigin:   defaultOrigin,
			DefaultTargets:  defaultTargets,
			OnUniverseClick: func() { setUniverseSelectView(controller) },
			OnAuthClick:     func() { setApiConfigurationView(controller) },
			OnHistoryClick: func() {
				dir, err := config.GetProgramDir()
				if err != nil {
					slog.Warn("failed to get program dir", "error", err)
					return
				}
				historyDir := filepath.Join(dir, "history", fmt.Sprint(universeId))
				os.MkdirAll(historyDir, 0755)

				u, err := url.Parse("file:///" + filepath.ToSlash(historyDir))
				if err == nil {
					fyne.CurrentApp().OpenURL(u)
				}
			},
			OnFileUploadClick: func() {
				go func() {
					path, err := sqdialog.File().Filter("Text files", "txt").Load()
					if err != nil {
						if err.Error() == "Cancelled" {
							return
						}
						fyne.Do(func() {
							dialog.ShowError(err, mainWindow)
						})
						return
					}

					copyStr, removeStr, err := controller.ParseTXTFile(path)
					if err != nil {
						fyne.Do(func() {
							dialog.ShowError(err, mainWindow)
						})
						return
					}

					if publishViewCfg.SetInputText != nil {
						fyne.Do(func() {
							publishViewCfg.SetInputText(copyStr, removeStr)
						})
					}
				}()
			},
			OnPublish: func(selectedOrigin views.Place, selectedTargets []views.Place, copy, remove string) {
				controller.ClearMemory()
				startPublishProcess(controller, universeId, selectedOrigin, selectedTargets, copy, remove)
			},
		}
		publishViewCfg.OnSettingsClick = func() {
			setUniConfigurationView(controller, universeId, publishViewCfg)
		}

		go func() {
			thumbnailURL, err := controller.Api.GetUniverseThumbnail(universeId, client.SMALL_THUMBNAIL)
			if err != nil {
				slog.Warn("failed to retrieve universe thumbnail", "error", err)
				return
			}
			publishViewCfg.ThumbnailURL = thumbnailURL
			fyne.Do(func() {
				publishViewCfg.SetThumbnail(thumbnailURL)
			})
		}()

		controller.SetLatestUniverse(universeId)
		publishView := views.PublishView(publishViewCfg)

		fyne.Do(func() {
			loadingBar.Stop()
			wrapper.Objects = []fyne.CanvasObject{publishView}
			mainWindow.SetTitle(fmt.Sprintf("Place publishing in %d - rbxpmux", universeId))
			wrapper.Refresh()
		})
	}()
}

func setUniConfigurationView(controller *controller.Controller, universeId int, pblCfg *views.PublishViewConfiguration) {
	uniCfg := &views.UniverseConfigViewConfiguration{
		Places:         pblCfg.Places,
		DefaultOrigin:  pblCfg.DefaultOrigin,
		DefaultTargets: pblCfg.DefaultTargets,
		OnSave: func(originId int, targetIds []int) {
			controller.SaveUniverseConfig(universeId, originId, targetIds)
			setPublishView(controller, universeId, pblCfg)
		},
		OnCancel: func() {
			setPublishView(controller, universeId, pblCfg)
		},
	}
	uniCfgView := views.UniverseConfigView(uniCfg)
	mainWindow.SetContent(uniCfgView)
	mainWindow.SetTitle(fmt.Sprintf("Universe %d configuration - rbxpmux", universeId))
}
