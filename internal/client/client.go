package client

import (
	"net/http"
	"time"
)

type RbxApi struct {
	key    string
	client *http.Client
}

func NewRbxApi(apiKey string) *RbxApi {
	return &RbxApi{
		key: apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (r *RbxApi) SetApiKey(apiKey string) {
	r.key = apiKey
}
