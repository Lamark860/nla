package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	apiKey  string
	baseURL string
	model   string
	proxy   string
	http    *http.Client
}

type Config struct {
	APIKey  string
	BaseURL string
	Model   string
	Proxy   string
	Timeout time.Duration
}

func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1/"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 200 * time.Second
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/") + "/"

	httpClient := &http.Client{Timeout: cfg.Timeout}

	// Configure proxy if provided
	if cfg.Proxy != "" {
		proxyURL, err := url.Parse("http://" + cfg.Proxy)
		if err == nil {
			httpClient.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	}

	return &Client{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
		proxy:   cfg.Proxy,
		http:    httpClient,
	}
}

// Model returns the model name the client was configured with. Used by
// callers that want to tag persisted LLM output with the producing model.
func (c *Client) Model() string {
	return c.model
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model               string        `json:"model"`
	Messages            []ChatMessage `json:"messages"`
	Temperature         *float64      `json:"temperature,omitempty"`
	MaxTokens           *int          `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int          `json:"max_completion_tokens,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ChatCompletion sends messages to OpenAI and returns the assistant's response.
// Retries up to 3 times on 429/5xx with exponential backoff.
func (c *Client) ChatCompletion(ctx context.Context, messages []ChatMessage) (string, error) {
	req := chatRequest{
		Model:    c.model,
		Messages: messages,
	}

	// Reasoning models (gpt-5.x, o1, o3, o4) use max_completion_tokens
	if isReasoningModel(c.model) {
		tokens := 5000
		req.MaxCompletionTokens = &tokens
	} else {
		temp := 0.3
		tokens := 5000
		req.Temperature = &temp
		req.MaxTokens = &tokens
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		result, statusCode, err := c.doRequest(ctx, "chat/completions", body)
		if err != nil {
			lastErr = err
			if attempt < maxRetries {
				time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * time.Second)
			}
			continue
		}

		if statusCode == 429 || statusCode >= 500 {
			lastErr = fmt.Errorf("openai returned %d", statusCode)
			if attempt < maxRetries {
				time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * time.Second)
			}
			continue
		}

		if statusCode != 200 {
			return "", fmt.Errorf("openai returned %d: %s", statusCode, string(result))
		}

		var resp chatResponse
		if err := json.Unmarshal(result, &resp); err != nil {
			return "", fmt.Errorf("decode response: %w", err)
		}

		if resp.Error != nil {
			return "", fmt.Errorf("openai error: %s", resp.Error.Message)
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("no choices in response")
		}

		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *Client) doRequest(ctx context.Context, endpoint string, body []byte) ([]byte, int, error) {
	url := c.baseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "nla-ai-client/1.0")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return data, resp.StatusCode, nil
}

func isReasoningModel(model string) bool {
	return strings.HasPrefix(model, "gpt-5") ||
		strings.HasPrefix(model, "o1") ||
		strings.HasPrefix(model, "o3") ||
		strings.HasPrefix(model, "o4")
}
