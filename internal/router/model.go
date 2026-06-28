package router

import (
	"encoding/json"
	"fmt"
	"strings"

	"api-in-one/internal/config"
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
	modelValue, err := modelFromPayload(payload)
	if err != nil {
		return "", nil, err
	}
	return rewritePayloadModel(payload, modelValue)
}

func RewriteModelBodyWithDefaults(body []byte, defaults config.DefaultsConfig) (string, []byte, error) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", nil, fmt.Errorf("invalid json body")
	}
	modelValue, err := modelFromPayload(payload)
	if err != nil {
		return "", nil, err
	}
	if strings.Contains(modelValue, "/") {
		return rewritePayloadModel(payload, modelValue)
	}
	if defaults.Provider == "" || defaults.Model == "" {
		return "", nil, fmt.Errorf("defaults.provider and defaults.model are required for unprefixed model routing")
	}
	modelRaw, err := json.Marshal(defaults.Model)
	if err != nil {
		return "", nil, fmt.Errorf("rewrite json body: %w", err)
	}
	payload["model"] = modelRaw
	rewritten, err := json.Marshal(payload)
	if err != nil {
		return "", nil, fmt.Errorf("rewrite json body: %w", err)
	}
	return defaults.Provider, rewritten, nil
}

func modelFromPayload(payload map[string]json.RawMessage) (string, error) {
	modelRaw, ok := payload["model"]
	if !ok {
		return "", fmt.Errorf("request body must include model")
	}
	var modelValue string
	if err := json.Unmarshal(modelRaw, &modelValue); err != nil || modelValue == "" {
		return "", fmt.Errorf("request body must include model")
	}
	return modelValue, nil
}

func rewritePayloadModel(payload map[string]json.RawMessage, modelValue string) (string, []byte, error) {
	provider, upstreamModel, err := ParseModelPrefix(modelValue)
	if err != nil {
		return "", nil, err
	}
	modelRaw, err := json.Marshal(upstreamModel)
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
