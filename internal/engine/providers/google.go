package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GoogleProvider implements the ModelProvider interface for Google Gemini
type GoogleProvider struct {
	apiKey string
	client *http.Client
}

// NewGoogleProvider creates a new Google provider
func NewGoogleProvider(apiKey string) *GoogleProvider {
	return &GoogleProvider{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// Name returns the provider name
func (p *GoogleProvider) Name() string {
	return "google"
}

// ListModels returns available Google models
func (p *GoogleProvider) ListModels(ctx context.Context) ([]string, error) {
	// Return known Gemini models
	return []string{
		"gemini-2.0-flash-exp",
		"gemini-1.5-pro",
		"gemini-1.5-flash",
		"gemini-1.0-pro",
	}, nil
}

// Request sends a non-streaming request
func (p *GoogleProvider) Request(ctx context.Context, prompt string, opts RequestOptions) (string, error) {
	contents := []map[string]interface{}{
		{
			"parts": []map[string]string{
				{"text": prompt},
			},
		},
	}

	reqBody := map[string]interface{}{
		"contents": contents,
		"generationConfig": map[string]interface{}{
			"temperature":     opts.Temperature,
			"maxOutputTokens": opts.MaxTokens,
		},
	}

	if opts.SystemPrompt != "" {
		reqBody["systemInstruction"] = map[string]interface{}{
			"parts": []map[string]string{
				{"text": opts.SystemPrompt},
			},
		}
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", opts.Model, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("google API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Google")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}

// Stream sends a streaming request
func (p *GoogleProvider) Stream(ctx context.Context, prompt string, opts RequestOptions) (<-chan string, <-chan error) {
	tokenChan := make(chan string, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(tokenChan)
		defer close(errChan)

		contents := []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		}

		reqBody := map[string]interface{}{
			"contents": contents,
			"generationConfig": map[string]interface{}{
				"temperature":     opts.Temperature,
				"maxOutputTokens": opts.MaxTokens,
			},
		}

		if opts.SystemPrompt != "" {
			reqBody["systemInstruction"] = map[string]interface{}{
				"parts": []map[string]string{
					{"text": opts.SystemPrompt},
				},
			}
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?key=%s&alt=sse", opts.Model, p.apiKey)
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			errChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := p.client.Do(req)
		if err != nil {
			errChan <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errChan <- fmt.Errorf("google API error (status %d): %s", resp.StatusCode, string(body))
			return
		}

		// Google uses SSE (Server-Sent Events) for streaming
		// For simplicity, fall back to non-streaming for now
		// Full SSE implementation would require more complex parsing
		response, err := p.Request(ctx, prompt, opts)
		if err != nil {
			errChan <- err
			return
		}

		select {
		case tokenChan <- response:
		case <-ctx.Done():
		}
	}()

	return tokenChan, errChan
}
