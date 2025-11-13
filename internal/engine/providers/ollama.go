package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// OllamaProvider implements the ModelProvider interface for Ollama
type OllamaProvider struct {
	baseURL string
	client  *http.Client
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(baseURL string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434" // Default Ollama endpoint
	}

	return &OllamaProvider{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// Name returns the provider name
func (p *OllamaProvider) Name() string {
	return "ollama"
}

// ListModels returns available models from `ollama list`
func (p *OllamaProvider) ListModels(ctx context.Context) ([]string, error) {
	// Execute `ollama list` command
	cmd := exec.CommandContext(ctx, "ollama", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run 'ollama list': %w", err)
	}

	// Parse output
	models := []string{}
	scanner := bufio.NewScanner(bytes.NewReader(output))

	// Skip header line
	if scanner.Scan() {
		// Header: NAME    ID    SIZE    MODIFIED
	}

	// Parse model lines
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Extract model name (first column)
		fields := strings.Fields(line)
		if len(fields) > 0 {
			modelName := fields[0]
			models = append(models, modelName)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse ollama list output: %w", err)
	}

	return models, nil
}

// Request sends a non-streaming request
func (p *OllamaProvider) Request(ctx context.Context, prompt string, opts RequestOptions) (string, error) {
	reqBody := map[string]interface{}{
		"model":  opts.Model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": opts.Temperature,
			"num_predict": opts.MaxTokens,
		},
	}

	if opts.SystemPrompt != "" {
		reqBody["system"] = opts.SystemPrompt
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
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
		return "", fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Response string `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Response, nil
}

// Stream sends a streaming request
func (p *OllamaProvider) Stream(ctx context.Context, prompt string, opts RequestOptions) (<-chan string, <-chan error) {
	tokenChan := make(chan string, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(tokenChan)
		defer close(errChan)

		reqBody := map[string]interface{}{
			"model":  opts.Model,
			"prompt": prompt,
			"stream": true,
			"options": map[string]interface{}{
				"temperature": opts.Temperature,
				"num_predict": opts.MaxTokens,
			},
		}

		if opts.SystemPrompt != "" {
			reqBody["system"] = opts.SystemPrompt
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
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
			errChan <- fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
			return
		}

		// Read streaming response
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			var chunk struct {
				Response string `json:"response"`
				Done     bool   `json:"done"`
			}

			if err := json.Unmarshal(scanner.Bytes(), &chunk); err != nil {
				errChan <- fmt.Errorf("failed to decode chunk: %w", err)
				return
			}

			if chunk.Response != "" {
				select {
				case tokenChan <- chunk.Response:
				case <-ctx.Done():
					return
				}
			}

			if chunk.Done {
				break
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("stream reading error: %w", err)
		}
	}()

	return tokenChan, errChan
}
