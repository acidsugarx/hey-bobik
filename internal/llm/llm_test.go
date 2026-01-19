package llm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"response":"Test note content", "done": true}`)
	}))
	defer ts.Close()

	client := &Client{
		BaseURL: ts.URL,
		Model:   "test-model",
	}

	resp, err := client.Generate(context.Background(), "System prompt", "User prompt")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp != "Test note content" {
		t.Errorf("expected 'Test note content', got %s", resp)
	}
}

func TestGenerateError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "internal server error")
	}))
	defer ts.Close()

	client := &Client{
		BaseURL: ts.URL,
		Model:   "test-model",
	}

	_, err := client.Generate(context.Background(), "System", "Prompt")
	if err == nil {
		t.Error("expected error for 500 status code, got nil")
	}
}

func TestGenerateDecodeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `invalid json`)
	}))
	defer ts.Close()

	client := &Client{
		BaseURL: ts.URL,
		Model:   "test-model",
	}

	_, err := client.Generate(context.Background(), "System", "Prompt")
	if err == nil {
		t.Error("expected error for invalid json, got nil")
	}
}
