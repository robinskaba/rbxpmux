package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func (api *RbxApi) DownloadPlace(placeId int) ([]byte, error) {
	latestVersion, err := api.fetchLatestVersion(placeId)
	if err != nil {
		return nil, fmt.Errorf("latest version fetch issue: %w", err)
	}
	cdn, err := api.fetchContentUrl(placeId, latestVersion)
	if err != nil {
		return nil, fmt.Errorf("cdn place url fetch issue: %w", err)
	}
	binary, err := api.getPlace(cdn)
	if err != nil {
		return nil, fmt.Errorf("place download issue: %w", err)
	}
	return binary, nil
}

func (api *RbxApi) fetchLatestVersion(placeId int) (int, error) {
	endpoint := fmt.Sprintf("https://apis.roblox.com/place-version-history-api/v1/%d/history", placeId)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Add("x-api-key", api.key)
	res, err := api.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	type versionResponse struct {
		Version string `json:"version"`
	}
	type historyApiResponse struct {
		PlaceVersions []versionResponse `json:"placeVersions"`
	}
	var fetchContent historyApiResponse
	err = json.NewDecoder(res.Body).Decode(&fetchContent)
	if err != nil {
		return 0, err
	}

	if len(fetchContent.PlaceVersions) == 0 {
		return 0, fmt.Errorf("missing place versions in response")
	}

	version, err := strconv.Atoi(fetchContent.PlaceVersions[0].Version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

func (api *RbxApi) fetchContentUrl(placeId int, version int) (string, error) {
	type urlFetchResponse struct {
		Location string `json:"location"`
	}

	endpoint := fmt.Sprintf("https://apis.roblox.com/asset-delivery-api/v1/assetId/%d/version/%d", placeId, version)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("x-api-key", api.key)

	res, err := api.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("url fetch resulted in %d", res.StatusCode)
	}

	var fetchContent urlFetchResponse
	err = json.NewDecoder(res.Body).Decode(&fetchContent)
	if err != nil {
		return "", err
	}
	return fetchContent.Location, nil
}

func (api *RbxApi) getPlace(cdn string) ([]byte, error) {
	res, err := api.client.Get(cdn)
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()

	binary, err := io.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}
	return binary, err
}
