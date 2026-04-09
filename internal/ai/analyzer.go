package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// anthropicRequest defines the structure of a request to the Anthropic API.
type anthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []message `json:"messages"`
}

// message represents a single message in the conversation with Claude.
type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnalysisResult holds the structured output from Claude's analysis of a journal entry.
type AnalysisResult struct {
	Mood     string // A single word or short phrase describing the emotional tone
	Analysis string // A thoughtful paragraph of insights about the entry
}

// AnalyzeEntry sends a journal entry to the Claude API and returns a mood and analysis.
// This is called automatically whenever a user creates a new journal entry.
func AnalyzeEntry(apiKey, title, body string) (*AnalysisResult, error) {
	// Build the prompt — we ask Claude to respond in a specific format
	// so we can reliably parse the mood and analysis separately
	prompt := fmt.Sprintf(`You are a compassionate journaling assistant. Analyze the following journal entry and respond in exactly this format:

MOOD: [single word or short phrase describing the emotional tone]
ANALYSIS: [2-3 sentences of thoughtful, empathetic insight about the entry]

Journal Entry Title: %s
Journal Entry: %s`, title, body)

	// Build the request payload
	reqBody := anthropicRequest{
		Model:     "claude-opus-4-5",
		MaxTokens: 300,
		Messages: []message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	// Marshal the request into JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the HTTP request to the Anthropic API
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers for the Anthropic API
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse the response from the Anthropic API
	var anthropicResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(anthropicResp.Content) == 0 {
		return nil, fmt.Errorf("empty response from Anthropic API")
	}

	// Parse Claude's response into mood and analysis
	return parseAnalysis(anthropicResp.Content[0].Text)
}

// parseAnalysis extracts the mood and analysis from Claude's formatted response.
// Claude is prompted to respond in a specific format so we can reliably split it.
func parseAnalysis(response string) (*AnalysisResult, error) {
	lines := strings.Split(response, "\n")

	result := &AnalysisResult{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "MOOD:") {
			// Extract everything after "MOOD: "
			result.Mood = strings.TrimSpace(strings.TrimPrefix(line, "MOOD:"))
		} else if strings.HasPrefix(line, "ANALYSIS:") {
			// Extract everything after "ANALYSIS: "
			result.Analysis = strings.TrimSpace(strings.TrimPrefix(line, "ANALYSIS:"))
		}
	}

	if result.Mood == "" || result.Analysis == "" {
		return nil, fmt.Errorf("failed to parse analysis response: %s", response)
	}

	return result, nil
}
