package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

type orModel struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type orResponse struct {
	Data []orModel `json:"data"`
}

var orCategories = []string{
	"programming", "roleplay", "marketing", "marketing/seo", "technology",
	"science", "translation", "legal", "finance", "health", "trivia", "academia",
}

var orInputModalities = []string{
	"text", "image", "audio", "file", "video",
}

var orSupportedParams = []string{
	"temperature", "top_p", "top_k", "min_p", "top_a", "frequency_penalty",
	"presence_penalty", "repetition_penalty", "max_tokens", "max_completion_tokens",
	"logit_bias", "logprobs", "top_logprobs", "prediction", "seed",
	"response_format", "structured_outputs", "stop", "tools", "tool_choice",
	"parallel_tool_calls", "include_reasoning", "reasoning", "reasoning_effort",
	"web_search_options", "verbosity",
}

func FetchOpenRouterModels() ([]string, string, error) {
	dimension := rand.Intn(3)
	var paramName, paramValue string

	switch dimension {
	case 0:
		paramName = "category"
		paramValue = orCategories[rand.Intn(len(orCategories))]
	case 1:
		paramName = "input_modalities"
		paramValue = orInputModalities[rand.Intn(len(orInputModalities))]
	case 2:
		paramName = "supported_parameters"
		paramValue = orSupportedParams[rand.Intn(len(orSupportedParams))]
	}

	source := fmt.Sprintf("openrouter(%s:%s)", paramName, paramValue)
	url := fmt.Sprintf("https://openrouter.ai/api/v1/models?%s=%s&limit=100", paramName, paramValue)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, source, fmt.Errorf("openrouter: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, source, fmt.Errorf("openrouter: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, source, fmt.Errorf("openrouter: %w", err)
	}

	var result orResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, source, fmt.Errorf("openrouter: %w", err)
	}

	var names []string
	seen := make(map[string]bool)
	for _, m := range result.Data {
		name := m.Name
		if name == "" {
			name = m.ID
		}
		name = trimModelName(name)
		if name != "" && !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
	}

	return names, source, nil
}

func trimModelName(name string) string {
	maxLen := 60
	if len(name) > maxLen {
		name = name[:maxLen]
	}
	return name
}