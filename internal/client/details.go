package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Place struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (api *RbxApi) GetPlaces(universeId int) ([]Place, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://develop.roblox.com/v1/universes/%d/places", universeId), nil)
	if err != nil {
		return nil, nil
	}
	q := req.URL.Query()
	q.Add("limit", "100") // try to get max, currently ignoring possible universes with 100+ places
	req.URL.RawQuery = q.Encode()

	res, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get places resulted in: %d", res.StatusCode)
	}

	type responseBody struct {
		Data []Place `json:"data"`
	}
	var content responseBody
	err = json.NewDecoder(res.Body).Decode(&content)
	if err != nil {
		return nil, err
	}

	return content.Data, nil
}

type Universe struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (api *RbxApi) GetUniverses(userId int) ([]Universe, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://games.roblox.com/v2/users/%d/games", userId), nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("limit", "50") // try to get max as well
	req.URL.RawQuery = q.Encode()

	res, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	type responseBody struct {
		Data []Universe `json:"data"`
	}
	var content responseBody
	err = json.NewDecoder(res.Body).Decode(&content)
	if err != nil {
		return nil, err
	}
	return content.Data, nil
}

type ThumbnailSize string

const (
	MINI_THUMBNAIL    ThumbnailSize = "50x50"
	SMALL_THUMBNAIL   ThumbnailSize = "128x128"
	MEDIUM_THUMBNAIL  ThumbnailSize = "150x150"
	LARGE_THUMBNAIL   ThumbnailSize = "256x256"
	EXTREME_THUMBNAIL ThumbnailSize = "512x512"
)

func (api *RbxApi) GetUniverseThumbnail(universeId int, size ThumbnailSize) (string, error) {
	res, err := api.client.Get(fmt.Sprintf("https://thumbnails.roblox.com/v1/games/icons?universeIds=%d&size=%s&format=Png&isCircular=false", universeId, size))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	type responseBody struct {
		Data []struct {
			Url string `json:"imageUrl"`
		} `json:"data"`
	}
	var content responseBody
	err = json.NewDecoder(res.Body).Decode(&content)
	if err != nil {
		return "", err
	}
	if len(content.Data) != 1 {
		return "", fmt.Errorf("unexpected amount of universes returned: %d", len(content.Data))
	}
	return content.Data[0].Url, nil
}
