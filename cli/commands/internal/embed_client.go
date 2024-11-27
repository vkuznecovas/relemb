package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func joinURL(basePath, endpoint string) string {
	basePath = strings.TrimSuffix(basePath, "/")
	endpoint = strings.TrimPrefix(endpoint, "/")
	return basePath + "/" + endpoint
}

type EmbedClient struct {
	endpoint string
	token    string
	c        http.Client
}

func NewEmbedClient(endpoint, token string) *EmbedClient {
	return &EmbedClient{
		c: http.Client{
			Timeout: time.Minute,
		},
		endpoint: endpoint,
		token:    token,
	}
}

type EmbeddingRequest struct {
	Text string `json:"text"`
}

func (ec EmbedClient) GetEmbedding(ctx context.Context, text []byte) ([]float64, error) {
	bd := EmbeddingRequest{
		Text: string(text),
	}

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(bd)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, joinURL(ec.endpoint, "embedding"), buf)
	if err != nil {
		return nil, err
	}

	if ec.token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", ec.token))
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := ec.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received unknown status: %v", res.StatusCode)
	}

	r := []float64{}
	err = json.NewDecoder(res.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	return r, nil
}
