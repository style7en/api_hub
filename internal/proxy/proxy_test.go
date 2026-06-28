package proxy

import (
	"bufio"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"api_hub/internal/config"
)

func TestForwardSendsProviderKeyAndBody(t *testing.T) {
	var gotPath string
	var gotAuth string
	var gotBody string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.String()
		gotAuth = r.Header.Get("Authorization")
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		gotBody = string(data)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer upstream.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions?x=1", strings.NewReader(`{"model":"gpt-4o"}`))
	rr := httptest.NewRecorder()

	err := Forward(rr, req, config.ProviderConfig{BaseURL: upstream.URL + "/v1", APIKey: "upstream-key"}, []byte(`{"model":"gpt-4o"}`))
	if err != nil {
		t.Fatalf("Forward returned error: %v", err)
	}
	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d", rr.Code)
	}
	if gotPath != "/v1/chat/completions?x=1" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotAuth != "Bearer upstream-key" {
		t.Fatalf("auth = %q", gotAuth)
	}
	if gotBody != `{"model":"gpt-4o"}` {
		t.Fatalf("body = %q", gotBody)
	}
	if rr.Body.String() != `{"ok":true}` {
		t.Fatalf("response body = %q", rr.Body.String())
	}
}

func TestForwardTreatsNestedBaseURLAsAPIRoot(t *testing.T) {
	var gotPath string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.String()
		_, _ = w.Write([]byte(`{}`))
	}))
	defer upstream.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()

	err := Forward(rr, req, config.ProviderConfig{BaseURL: upstream.URL + "/proxy/openai/v1", APIKey: "upstream-key"}, []byte(`{}`))
	if err != nil {
		t.Fatalf("Forward returned error: %v", err)
	}
	if gotPath != "/proxy/openai/v1/chat/completions" {
		t.Fatalf("path = %q", gotPath)
	}
}

func TestForwardFlushesStreamingChunks(t *testing.T) {
	continueUpstream := make(chan struct{})
	var once sync.Once
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"delta\":\"hi\"}\n\n"))
		w.(http.Flusher).Flush()
		<-continueUpstream
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		w.(http.Flusher).Flush()
	}))
	defer upstream.Close()
	defer once.Do(func() { close(continueUpstream) })

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := Forward(w, r, config.ProviderConfig{BaseURL: upstream.URL + "/v1", APIKey: "upstream-key"}, []byte(`{}`)); err != nil {
			t.Errorf("Forward returned error: %v", err)
		}
	})
	proxyServer := httptest.NewServer(handler)
	defer proxyServer.Close()

	client := proxyServer.Client()
	respCh := make(chan *http.Response, 1)
	errCh := make(chan error, 1)
	go func() {
		resp, err := client.Post(proxyServer.URL+"/v1/chat/completions", "application/json", strings.NewReader(`{}`))
		if err != nil {
			errCh <- err
			return
		}
		respCh <- resp
	}()

	select {
	case resp := <-respCh:
		defer resp.Body.Close()
		lineCh := make(chan string, 1)
		readErrCh := make(chan error, 1)
		go func() {
			line, err := bufio.NewReader(resp.Body).ReadString('\n')
			if err != nil {
				readErrCh <- err
				return
			}
			lineCh <- line
		}()
		select {
		case line := <-lineCh:
			if line != "data: {\"delta\":\"hi\"}\n" {
				t.Fatalf("first line = %q", line)
			}
			once.Do(func() { close(continueUpstream) })
		case err := <-readErrCh:
			t.Fatalf("read first line: %v", err)
		case <-time.After(500 * time.Millisecond):
			t.Fatal("timed out waiting for first streamed event")
		}
	case err := <-errCh:
		t.Fatal(err)
	case <-time.After(500 * time.Millisecond):
		once.Do(func() { close(continueUpstream) })
		t.Fatal("timed out waiting for streaming response headers")
	}
}

func TestForwardReturnsErrorOnNetworkFailure(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()

	err := Forward(rr, req, config.ProviderConfig{BaseURL: "http://127.0.0.1:1/v1", APIKey: "upstream-key"}, []byte(`{}`))
	if err == nil {
		t.Fatal("expected error")
	}
}
