package router

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ParseModelPrefix(model string) (string, string, error) {
	provider, upstreamModel, ok := strings.Cut(model, "/")
	if !ok || provider == "" || upstreamModel == "" || strings.TrimSpace(provider) != provider || strings.TrimSpace(upstreamModel) != upstreamModel {
		return "", "", fmt.Errorf("model must use provider/model format")
	}
	return provider, upstreamModel, nil
}

func RewriteModelBody(body []byte) (string, []byte, error) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", nil, fmt.Errorf("invalid json body")
	}
	modelRaw, ok := payload["model"]
	if !ok {
		return "", nil, fmt.Errorf("request body must include model")
	}
	var modelValue string
	if err := json.Unmarshal(modelRaw, &modelValue); err != nil || modelValue == "" {
		return "", nil, fmt.Errorf("request body must include model")
	}
	provider, upstreamModel, err := ParseModelPrefix(modelValue)
	if err != nil {
		return "", nil, err
	}
	modelRaw, err = json.Marshal(upstreamModel)
	if err != nil {
		return "", nil, fmt.Errorf("rewrite json body: %w", err)
	}
	payload["model"] = modelRaw
	rewritten, err := json.Marshal(payload)
	if err != nil {
		return "", nil, fmt.Errorf("rewrite json body: %w", err)
	}
	return provider, rewritten, nil
}
