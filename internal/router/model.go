package router

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ParseModelPrefix(model string) (string, string, error) {
	provider, upstreamModel, ok := strings.Cut(model, "/")
	if !ok || provider == "" || upstreamModel == "" {
		return "", "", fmt.Errorf("model must use provider/model format")
	}
	return provider, upstreamModel, nil
}

func RewriteModelBody(body []byte) (string, []byte, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", nil, fmt.Errorf("invalid json body")
	}
	modelValue, ok := payload["model"].(string)
	if !ok || modelValue == "" {
		return "", nil, fmt.Errorf("request body must include model")
	}
	provider, upstreamModel, err := ParseModelPrefix(modelValue)
	if err != nil {
		return "", nil, err
	}
	payload["model"] = upstreamModel
	rewritten, err := json.Marshal(payload)
	if err != nil {
		return "", nil, fmt.Errorf("rewrite json body: %w", err)
	}
	return provider, rewritten, nil
}
