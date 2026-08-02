package config

import (
	"errors"
	"strconv"

	"github.com/zalando/go-keyring"
)

type Settings struct {
	UserId int
	ApiKey string
}

const service = "rbxpmux"

func LoadSettings() (*Settings, error) {
	var settings Settings
	userIdStr, err := keyring.Get(service, "user-id")
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return nil, err
	} else if !errors.Is(err, keyring.ErrNotFound) {
		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			return nil, err
		}
		settings.UserId = userId
	}
	apiKey, err := keyring.Get(service, "api-key")
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return nil, err
	}
	settings.ApiKey = apiKey
	return &settings, nil
}

func (settings *Settings) Save() error {
	err := keyring.Set(service, "user-id", strconv.Itoa(settings.UserId))
	if err != nil {
		return err
	}
	err = keyring.Set(service, "api-key", settings.ApiKey)
	if err != nil {
		return err
	}
	return nil
}
