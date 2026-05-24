package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type HuggingFaceProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewHuggingFaceProvider(apiKey, model string) *HuggingFaceProvider {
	return &HuggingFaceProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}
}

type hfChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type hfChatRequest struct {
	Model       string           `json:"model"`
	Messages    []hfChatMessage  `json:"messages"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Temperature float64          `json:"temperature,omitempty"`
	Stream      bool             `json:"stream,omitempty"`
}

type hfChatChoice struct {
	Index   int `json:"index"`
	Message struct {
		Role      string `json:"role"`
		Content   string `json:"content"`
		Reasoning string `json:"reasoning"`
	} `json:"message"`
}

type hfChatResponse struct {
	Choices []hfChatChoice `json:"choices"`
}

func (p *HuggingFaceProvider) GenerateJSON(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	reqBody := hfChatRequest{
		Model: p.model,
		Messages: []hfChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt + "\n\nRespond only with valid JSON."},
		},
		MaxTokens:   8192,
		Temperature: 0.3,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	apiURL := "https://router.huggingface.co/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("huggingface api: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("huggingface api status %d: %s", resp.StatusCode, string(respBody))
	}

	var result hfChatResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse response (%d bytes): %w. body: %s", len(respBody), err, string(respBody))
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("empty huggingface response")
	}

	msg := result.Choices[0].Message
	text := msg.Content
	if text == "" {
		text = msg.Reasoning
	}
	return text, nil
}

func (p *HuggingFaceProvider) GenerateStream(ctx context.Context, systemPrompt, userPrompt string, onToken func(string)) error {
	reqBody := hfChatRequest{
		Model: p.model,
		Messages: []hfChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens:   8192,
		Temperature: 0.7,
		Stream:      true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	apiURL := "https://router.huggingface.co/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("huggingface api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("huggingface api status %d: %s", resp.StatusCode, string(respBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		if data == "[DONE]" {
			break
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content   string `json:"content"`
					Reasoning string `json:"reasoning"`
				} `json:"delta"`
			} `json:"choices"`
		}

		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 {
			text := chunk.Choices[0].Delta.Content
			if text == "" {
				text = chunk.Choices[0].Delta.Reasoning
			}
			if text != "" {
				onToken(text)
			}
		}
	}

	return scanner.Err()
}
