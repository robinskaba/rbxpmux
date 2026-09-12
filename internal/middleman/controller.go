package middleman

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/robinskaba/rbxpmux/internal/client"
	"github.com/robinskaba/rbxpmux/internal/config"
	"github.com/robinskaba/rbxpmux/internal/editor"
)

type Controller struct {
	Api           *client.RbxApi
	Settings      *config.Settings
	Configuration *config.PMUXConfig
	memory        editMemory
}

type editMemory struct {
	sync.Mutex
	origin  []byte
	targets map[int][]byte
	results map[int][]byte
}

var ErrFailedToSetNotes error = fmt.Errorf("publishing notes failed")

func NewController(api *client.RbxApi, settings *config.Settings, configuration *config.PMUXConfig) *Controller {
	return &Controller{
		Api:           api,
		Settings:      settings,
		Configuration: configuration,
		memory: editMemory{
			origin:  []byte{},
			targets: map[int][]byte{},
		},
	}
}

func (c *Controller) SetApiKey(apiKey string) {
	c.Settings.ApiKey = apiKey
	c.Api.SetApiKey(apiKey)
}

func (c *Controller) SaveApiConfig(userId int, apiKey string) {
	c.Settings.UserId = userId
	c.SetApiKey(apiKey)
	go func() {
		c.Settings.Save()
	}()
}

func (c *Controller) SetLatestUniverse(universeId int) {
	go func() {
		c.Configuration.LatestUniverseId = universeId
		c.Configuration.Save()
	}()
}

func (c *Controller) GetUniverseDefaults(universeId int) (int, []int) {
	cfg, ok := c.Configuration.UniverseConfigs[universeId]
	if ok {
		return cfg.DefaultOriginId, cfg.DefaultTargetsIds
	}
	return 0, nil
}

func (c *Controller) SaveUniverseConfig(universeId int, originId int, targetIds []int) {
	c.Configuration.SetUniverseConfig(universeId, originId, targetIds)
	c.Configuration.Save()
}

// api
func (c *Controller) ClearMemory() {
	c.memory.Lock()
	defer c.memory.Unlock()
	c.memory.origin = []byte{}
	c.memory.targets = make(map[int][]byte)
	c.memory.results = make(map[int][]byte)
}

func (c *Controller) DownloadOrigin(placeId int) error {
	slog.Info("downloading origin place", "placeId", placeId)
	binary, err := c.Api.DownloadPlace(placeId)
	if err != nil {
		return err
	}
	c.memory.origin = binary
	return nil
}
func (c *Controller) DownloadTarget(placeId int) error {
	slog.Info("downloading target place", "placeId", placeId)
	binary, err := c.Api.DownloadPlace(placeId)
	if err != nil {
		return err
	}
	c.memory.Lock()
	c.memory.targets[placeId] = binary
	c.memory.Unlock()
	return nil
}

func (c *Controller) EditPlaces(copy, remove string) error {
	var instructions []editor.Instruction
	txtToInstructions := func(txt string, insType editor.InstructionType) []editor.Instruction {
		txt = strings.ReplaceAll(txt, "\r\n", "\n")
		lines := strings.Split(txt, "\n")
		var parsed []editor.Instruction
		for _, line := range lines {
			line = strings.TrimSpace(line)

			// strip whitespace characters
			line = strings.ReplaceAll(line, "\t", "")

			// remove --comments from instruction (workspace.Items -- new item)
			commentIdx := strings.Index(line, "--")
			if commentIdx != -1 {
				line = line[:commentIdx]
			}
			line = strings.TrimRight(line, " ")

			if line == "" {
				continue
			}
			parsed = append(parsed, editor.Instruction{
				Type:    insType,
				Content: line,
			})
		}
		return parsed
	}
	instructions = append(instructions, txtToInstructions(copy, editor.COPY)...)
	instructions = append(instructions, txtToInstructions(remove, editor.REMOVE)...)

	slog.Info("editing places", "instructions_count", len(instructions))
	newPlaces, err := editor.EditPlaces(c.memory.origin, c.memory.targets, instructions)
	if err != nil {
		return err
	}
	c.memory.results = newPlaces
	return nil
}

type PublishPlaceParams struct {
	UniverseId int
	PlaceId    int
	SaveOnly   bool
	NoteTitle  string
	NoteDesc   string
}

func (c *Controller) PublishPlace(params PublishPlaceParams) error {
	converted, ok := c.memory.results[params.PlaceId]
	if !ok {
		return fmt.Errorf("converted place is missing in memory: %d", params.PlaceId)
	}

	pslog := slog.With("universeId", params.UniverseId, "placeId", params.PlaceId)
	var err error

	for attempt := 1; attempt <= 4; attempt++ {
		version, err := c.Api.PublishPlace(params.UniverseId, params.PlaceId, converted, params.SaveOnly)
		if err == nil {
			pslog.Info("publishing successful", "attempts", attempt)

			// set notes
			if params.NoteTitle == "" {
				params.NoteTitle = fmt.Sprintf("Version %d", version)
			}
			err = c.Api.SetVersionNotes(params.PlaceId, version, params.NoteTitle, params.NoteDesc)
			if err != nil {
				pslog.Error("setting version notes failed", "error", err)
				return ErrFailedToSetNotes // custom error to not stop the complete publishing flow
			}

			return nil
		}

		if attempt < 4 {
			pslog.Info("publishing failed, retrying", "attempt", attempt, "error", err)
			time.Sleep(2 * time.Second)
		}
	}

	return err
}

func (c *Controller) ParseTXTFile(path string) (string, string, error) {
	return parseTXTFile(path)
}

func (c *Controller) ArchiveTXTFile(universeId int, note, copy, remove string) error {
	dir, err := config.GetProgramDir()
	if err != nil {
		return err
	}

	historyDir := filepath.Join(dir, "history", fmt.Sprint(universeId))
	err = os.MkdirAll(historyDir, 0755)
	if err != nil {
		return err
	}

	now := time.Now().Format("02-01-2006_15-04")
	destPath := filepath.Join(historyDir, now+".txt")

	output := strings.Builder{}
	output.WriteString("-- ")
	output.WriteString(note)
	output.WriteString("\n\n")
	output.WriteString(mergeInputStrings(copy, remove))

	err = os.WriteFile(destPath, []byte(output.String()), 0644)
	if err == nil {
		slog.Info("instruction file archived", "path", destPath)
	}
	return err
}

type InternalDir string

const INTERNAL_DIR_HISTORY = "history"
const INTERNAL_DIR_LOGS = "logs"

func (c *Controller) GetInternalDirURL(internalDir InternalDir, universeId int) (*url.URL, error) {
	dir, err := config.GetProgramDir()
	if err != nil {
		slog.Warn("failed to get program dir", "error", err)
		return nil, err
	}

	var fullPath string
	switch internalDir {
	case INTERNAL_DIR_HISTORY:
		fullPath = filepath.Join(dir, string(internalDir), fmt.Sprint(universeId))
		err = os.MkdirAll(fullPath, 0755)
		if err != nil {
			slog.Warn("failed to create history dir", "error", err)
			return nil, err
		}
	case INTERNAL_DIR_LOGS:
		fullPath = filepath.Join(dir, string(internalDir))
	}

	if fullPath == "" {
		return nil, fmt.Errorf("invalid internal dir: %s", internalDir)
	}

	u, err := url.Parse("file:///" + filepath.ToSlash(fullPath))
	return u, err
}
