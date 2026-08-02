package controller

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

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

func (c *Controller) PublishPlace(universeId int, placeId int) error {
	converted, ok := c.memory.results[placeId]
	if !ok {
		return fmt.Errorf("converted place is missing in memory: %d", placeId)
	}
	slog.Info("publishing place", "universeId", universeId, "placeId", placeId)
	err := c.Api.PublishPlace(universeId, placeId, converted)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) ParseTXTFile(path string) (string, string, error) {
	return parseTXTFile(path)
}

func (c *Controller) ArchiveTXTFile(universeId int, copyStr, removeStr string) error {
	return archiveTXTFile(universeId, copyStr, removeStr)
}
