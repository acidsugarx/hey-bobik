package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultTimeout = 60 * time.Second
)

// Client handles communication with the Ollama API.
type Client struct {
	BaseURL    string
	Model      string
	Timeout    time.Duration
	httpClient *http.Client
}

// GenerateRequest represents the request body for Ollama's generate API.
type GenerateRequest struct {
	Model  string   `json:"model"`
	Prompt string   `json:"prompt"`
	System string   `json:"system"`
	Stream bool     `json:"stream"`
	Images []string `json:"images,omitempty"` // Для vision моделей: base64-encoded изображения
}

// GenerateResponse represents the response body from Ollama's generate API.
type GenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// New creates a new Ollama client.
func New(baseURL, model string) *Client {
	return &Client{
		BaseURL: baseURL,
		Model:   model,
		Timeout: defaultTimeout,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Generate sends a prompt to Ollama and returns the generated response.
func (c *Client) Generate(ctx context.Context, system, prompt string) (string, error) {
	return c.GenerateWithImages(ctx, system, prompt, nil)
}

// GenerateWithImages отправляет запрос к Ollama с поддержкой изображений (для vision моделей).
// images - список изображений в формате base64.
func (c *Client) GenerateWithImages(ctx context.Context, system, prompt string, images []string) (string, error) {
	reqBody := GenerateRequest{
		Model:  c.Model,
		Prompt: prompt,
		System: system,
		Stream: false,
		Images: images,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var genResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return genResp.Response, nil
}
