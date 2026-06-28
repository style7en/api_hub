package openaierror

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteReturnsOpenAIStyleError(t *testing.T) {
	rr := httptest.NewRecorder()

	Write(rr, http.StatusBadRequest, "invalid_request_error", "missing model prefix")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q", got)
	}
	var body map[string]map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["error"]["type"] != "invalid_request_error" {
		t.Fatalf("type = %q", body["error"]["type"])
	}
	if body["error"]["message"] != "missing model prefix" {
		t.Fatalf("message = %q", body["error"]["message"])
	}
}
