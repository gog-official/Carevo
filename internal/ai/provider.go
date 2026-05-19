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

type GeminiProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewGeminiProvider(apiKey, model string) *GeminiProvider {
	return &GeminiProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}
}

type geminiRequest struct {
	Contents         []geminiContent  `json:"contents"`
	GenerationConfig *generationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type generationConfig struct {
	ResponseMimeType string `json:"response_mime_type,omitempty"`
	Temperature      float64 `json:"temperature,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (p *GeminiProvider) GenerateJSON(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	fullPrompt := systemPrompt + "\n\n" + userPrompt

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: fullPrompt}}},
		},
		GenerationConfig: &generationConfig{
			ResponseMimeType: "application/json",
			Temperature:      0.3,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", p.model, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini api: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini api status %d: %s", resp.StatusCode, string(respBody))
	}

	var gr geminiResponse
	if err := json.Unmarshal(respBody, &gr); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty gemini response")
	}

	return gr.Candidates[0].Content.Parts[0].Text, nil
}

func (p *GeminiProvider) GenerateStream(ctx context.Context, systemPrompt, userPrompt string, onToken func(string)) error {
	fullPrompt := systemPrompt + "\n\n" + userPrompt
	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: fullPrompt}}},
		},
		GenerationConfig: &generationConfig{
			Temperature: 0.7,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", p.model, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("gemini api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gemini api status %d: %s", resp.StatusCode, string(respBody))
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

		var sr geminiResponse
		if err := json.Unmarshal([]byte(data), &sr); err != nil {
			continue
		}

		if len(sr.Candidates) > 0 && len(sr.Candidates[0].Content.Parts) > 0 {
			onToken(sr.Candidates[0].Content.Parts[0].Text)
		}
	}

	return scanner.Err()
}
