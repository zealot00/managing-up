package gateway

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleOpenAIChat_MissingModel(t *testing.T) {
	s := &Server{}

	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")

	w := httptest.NewRecorder()
	s.HandleOpenAIChat(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	errObj, ok := resp["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected error object in response")
	}

	if errObj["message"] != "model is required" {
		t.Errorf("expected error message 'model is required', got %v", errObj["message"])
	}
}

func TestHandleOpenAIChat_MissingAuth(t *testing.T) {
	s := &Server{}

	reqBody := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header

	w := httptest.NewRecorder()
	s.HandleOpenAIChat(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	errObj, ok := resp["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected error object in response")
	}

	if errObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got %v", errObj["code"])
	}
}

func TestHandleOpenAIChat_StreamNotSupported(t *testing.T) {
	s := &Server{}

	reqBody := map[string]interface{}{
		"model":  "gpt-4o",
		"stream": true,
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")

	w := httptest.NewRecorder()
	s.HandleOpenAIChat(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	errObj, ok := resp["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected error object in response")
	}

	if errObj["code"] != "stream_failed" {
		t.Errorf("expected error code 'stream_failed', got %v", errObj["code"])
	}
}

func TestHandleOpenAIChat_MethodNotAllowed(t *testing.T) {
	s := &Server{}

	req := httptest.NewRequest(http.MethodGet, "/v1/chat/completions", nil)

	w := httptest.NewRecorder()
	s.HandleOpenAIChat(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}
