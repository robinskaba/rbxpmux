package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type universeConfig struct {
	DefaultOriginId   int   `json:"default_origin_id"`
	DefaultTargetsIds []int `json:"default_targets_ids"`
}

type PMUXConfig struct {
	LatestUniverseId int                    `json:"latest_universe_id"`
	UniverseConfigs  map[int]universeConfig `json:"universe_configs"`
}

func getConfigPath() (string, error) {
	configDir, err := GetProgramDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

func LoadConfig() (*PMUXConfig, error) {
	var cfg PMUXConfig
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// no config set up yet, use default
		if os.IsNotExist(err) {
			return &cfg, nil
		}
		return nil, err
	}

	err = json.Unmarshal(data, &cfg)
	return &cfg, err
}

func (cfg *PMUXConfig) Save() error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600) // sets permission to only current user
}

func (cfg *PMUXConfig) SetUniverseConfig(universeId int, originId int, targetIds []int) {
	if cfg.UniverseConfigs == nil {
		cfg.UniverseConfigs = map[int]universeConfig{}
	}
	cfg.UniverseConfigs[universeId] = universeConfig{
		DefaultOriginId:   originId,
		DefaultTargetsIds: targetIds,
	}
}
