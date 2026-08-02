package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func (api *RbxApi) PublishPlace(universeId int, placeId int, file []byte) error {
	return api.postPlace(universeId, placeId, file)
}

var ErrPlaceOpenInStudio error = fmt.Errorf("conflict caused by having the place open in studio")

type unacceptableResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (api *RbxApi) postPlace(universeId int, placeId int, file []byte) error {
	endpoint := fmt.Sprintf("https://apis.roblox.com/universes/v1/%d/places/%d/versions", universeId, placeId) // limit: 30 requests / minute
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(file))
	if err != nil {
		return err
	}
	req.Header.Add("x-api-key", api.key)
	req.Header.Add("Content-Type", "application/octet-stream")

	query := req.URL.Query()
	query.Add("versionType", "Published") // 'Saved' = save only, 'Published' = save + publish
	req.URL.RawQuery = query.Encode()

	res, err := api.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		var unRes unacceptableResponse
		err = json.NewDecoder(res.Body).Decode(&unRes)
		if err != nil {
			return err
		}

		if unRes.Code == "Conflict" {
			return ErrPlaceOpenInStudio
		}
		return fmt.Errorf("place publish resulted in: %d (%s) - %s", res.StatusCode, unRes.Code, unRes.Message)
	}
	return nil
}
