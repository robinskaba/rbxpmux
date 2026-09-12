package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// sets the notes for a version of a place
func (api *RbxApi) SetVersionNotes(placeId int, version int, noteTitle, noteDesc string) error {
	endpoint := fmt.Sprintf("https://apis.roblox.com/place-version-history-api/v1/%d/version/%d/notes", placeId, version)

	payload := struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}{
		Title:       noteTitle,
		Description: noteDesc,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Add("x-api-key", api.key)
	req.Header.Add("Content-Type", "application/json")

	res, err := api.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode) // 500 code seems to mean roblox denied it - too long (max 1000 char) or doesnt meet community standards
	}

	return nil
}
