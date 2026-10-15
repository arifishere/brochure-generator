package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
)

const (
	ollamaGenerateURL = "http://localhost:11434/api/generate"
	ollamaModel       = "gemma4:e2b"

	systemPrompt = `You extract company facts for a brochure. Ignore navigation, ads, scripts, and markup. Reply with ONLY these lines, one per field, no intro or closing text:

Company Name: ...
Established: ...
Core Business: ...
Target Audience: ...
USP: ...
Description: ...

Description should be 2-4 sentences about what this page/website is about. If a value is unknown, use "Unknown". Keep other fields short and factual.`
	userPrefix = "From the following website content, fill the brochure fields:"
)

type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	System string `json:"system,omitempty"`
	Stream bool   `json:"stream"`
}

type generateResponse struct {
	Response string `json:"response"`
}

// AskOllama sends a prompt to the local Ollama generate API and returns the completion.
func AskOllama(model, system, prompt string) (string, error) {
	body, err := json.Marshal(generateRequest{
		Model:  model,
		Prompt: prompt,
		System: system,
		Stream: false,
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	resp, err := http.Post(ollamaGenerateURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("call ollama: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result generateResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	return result.Response, nil
}

// FetchWebsite fetches the content at url and returns it as a string.
func FetchWebsite(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetch website: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("website returned %d", resp.StatusCode)
	}

	return string(body), nil
}

var (
	scriptTagRe = regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script>`)
	styleTagRe  = regexp.MustCompile(`(?is)<style\b[^>]*>.*?</style>`)
	noscriptRe  = regexp.MustCompile(`(?is)<noscript\b[^>]*>.*?</noscript>`)
	htmlTagRe   = regexp.MustCompile(`(?s)<[^>]+>`)
	spaceRe     = regexp.MustCompile(`\s+`)
)

// StripMarkup removes script, style, noscript, and HTML tags, then returns plain text.
func StripMarkup(content string) string {
	content = scriptTagRe.ReplaceAllString(content, " ")
	content = styleTagRe.ReplaceAllString(content, " ")
	content = noscriptRe.ReplaceAllString(content, " ")
	content = htmlTagRe.ReplaceAllString(content, " ")
	content = html.UnescapeString(content)
	content = spaceRe.ReplaceAllString(content, " ")
	return strings.TrimSpace(content)
}

// SummarizeWebsite fetches url, strips markup, and summarizes with the local LLM.
func SummarizeWebsite(url string) (string, error) {
	content, err := FetchWebsite(url)
	if err != nil {
		return "", fmt.Errorf("summarize website: %w", err)
	}

	text := StripMarkup(content)
	prompt := userPrefix + "\n\n" + text

	summary, err := AskOllama(ollamaModel, systemPrompt, prompt)
	if err != nil {
		return "", fmt.Errorf("summarize website: %w", err)
	}

	return summary, nil
}
