package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

var ErrPlaceOpenInStudio error = fmt.Errorf("conflict caused by having the place open in studio")

type unacceptableResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (api *RbxApi) PublishPlace(universeId int, placeId int, file []byte, saveOnly bool) (int, error) {
	endpoint := fmt.Sprintf("https://apis.roblox.com/universes/v1/%d/places/%d/versions", universeId, placeId) // limit: 30 requests / minute
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(file))
	if err != nil {
		return 0, err
	}
	req.Header.Add("x-api-key", api.key)
	req.Header.Add("Content-Type", "application/octet-stream")

	query := req.URL.Query()

	// 'Saved' = save only, 'Published' = save + publish
	versionType := "Published"
	if saveOnly {
		versionType = "Saved"
	}
	query.Add("versionType", versionType)
	req.URL.RawQuery = query.Encode()

	res, err := api.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		var unRes unacceptableResponse
		err = json.NewDecoder(res.Body).Decode(&unRes)
		if err != nil {
			return 0, err
		}

		if unRes.Code == "Conflict" {
			return 0, ErrPlaceOpenInStudio
		}
		return 0, fmt.Errorf("place publish resulted in: %d (%s) - %s", res.StatusCode, unRes.Code, unRes.Message)
	}

	var content struct {
		VersionNumber int `json:"versionNumber"`
	}
	err = json.NewDecoder(res.Body).Decode(&content)
	if err != nil {
		return 0, fmt.Errorf("version response decoding error: %w", err)
	}

	return content.VersionNumber, nil
}
