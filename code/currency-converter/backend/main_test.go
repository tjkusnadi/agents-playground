package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConvertHandlerMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/convert?base=USD&target=IDR", nil)
	res := httptest.NewRecorder()

	convertHandler(res, req)

	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, res.Code)
	}
}

func TestConvertHandlerValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{
			name:       "missing base",
			url:        "/api/convert?target=IDR",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid amount",
			url:        "/api/convert?base=USD&target=IDR&amount=abc",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			res := httptest.NewRecorder()

			convertHandler(res, req)

			if res.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, res.Code)
			}
		})
	}
}

func TestConvertHandlerSuccess(t *testing.T) {
	originalFetcher := rateFetcher
	rateFetcher = func(base, target string) (float64, error) {
		if base != "USD" || target != "IDR" {
			t.Fatalf("unexpected arguments: %s, %s", base, target)
		}
		return 15000.5, nil
	}
	defer func() { rateFetcher = originalFetcher }()

	req := httptest.NewRequest(http.MethodGet, "/api/convert?base=USD&target=IDR&amount=2", nil)
	res := httptest.NewRecorder()

	convertHandler(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	var payload convertResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload.Rate != 15000.5 {
		t.Fatalf("expected rate 15000.5, got %f", payload.Rate)
	}

	if payload.Converted != 30001 {
		t.Fatalf("expected converted 30001, got %f", payload.Converted)
	}
}

func TestConvertHandlerFetchError(t *testing.T) {
	originalFetcher := rateFetcher
	rateFetcher = func(string, string) (float64, error) {
		return 0, errors.New("boom")
	}
	defer func() { rateFetcher = originalFetcher }()

	req := httptest.NewRequest(http.MethodGet, "/api/convert?base=USD&target=IDR", nil)
	res := httptest.NewRecorder()

	convertHandler(res, req)

	if res.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, res.Code)
	}
}

func TestWithCORSHandlesOptions(t *testing.T) {
	called := false
	handler := withCORS(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/convert", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, res.Code)
	}

	if called {
		t.Fatalf("expected handler not to be called on OPTIONS request")
	}

	allowOrigin := res.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin '*', got %q", allowOrigin)
	}
}

func TestWithCORSPassesThrough(t *testing.T) {
	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/convert", strings.NewReader(""))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	if res.Body.String() != "ok" {
		t.Fatalf("expected body 'ok', got %q", res.Body.String())
	}
}
