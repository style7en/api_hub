package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"api_hub/internal/config"
)

func Forward(w http.ResponseWriter, r *http.Request, provider config.ProviderConfig, body []byte) error {
	upstreamURL, err := joinURL(provider.BaseURL, r.URL.Path, r.URL.RawQuery)
	if err != nil {
		return err
	}
	upstreamReq, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create upstream request: %w", err)
	}
	copyRequestHeaders(upstreamReq.Header, r.Header)
	upstreamReq.Header.Set("Authorization", "Bearer "+provider.APIKey)
	upstreamReq.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))

	resp, err := http.DefaultClient.Do(upstreamReq)
	if err != nil {
		return fmt.Errorf("upstream request: %w", err)
	}
	defer resp.Body.Close()

	copyResponseHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	return copyResponseBody(w, resp.Body)
}

func copyResponseBody(w http.ResponseWriter, body io.Reader) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		_, _ = io.Copy(w, body)
		return nil
	}

	buf := make([]byte, 32*1024)
	for {
		n, err := body.Read(buf)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				return nil
			}
			flusher.Flush()
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return nil
		}
	}
}

func joinURL(baseURL string, requestPath string, rawQuery string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse provider base_url: %w", err)
	}
	basePath := strings.TrimRight(base.Path, "/")
	path := strings.TrimPrefix(requestPath, "/v1")
	if path == "" {
		path = "/"
	}
	base.Path = basePath + path
	base.RawQuery = rawQuery
	return base.String(), nil
}

func copyRequestHeaders(dst http.Header, src http.Header) {
	for key, values := range src {
		if strings.EqualFold(key, "Authorization") || strings.EqualFold(key, "Host") || strings.EqualFold(key, "Content-Length") {
			continue
		}
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func copyResponseHeaders(dst http.Header, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}
