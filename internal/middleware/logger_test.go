package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogger_PassesThrough(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello")) //nolint:errcheck
	})

	handler := Logger(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("Logger did not call next handler")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status: expected %d, got %d", http.StatusOK, w.Code)
	}
	if w.Body.String() != "hello" {
		t.Errorf("body: expected %q, got %q", "hello", w.Body.String())
	}
}

func TestLogger_CapturesStatusCode(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	handler := Logger(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTeapot {
		t.Errorf("status: expected %d, got %d", http.StatusTeapot, w.Code)
	}
}

func TestLogger_DefaultStatus200(t *testing.T) {
	// Next handler writes body without calling WriteHeader → default 200
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("body")) //nolint:errcheck
	})

	handler := Logger(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected default 200, got %d", w.Code)
	}
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec, status: http.StatusOK}

	rw.WriteHeader(http.StatusNotFound)

	if rw.status != http.StatusNotFound {
		t.Errorf("expected captured status %d, got %d", http.StatusNotFound, rw.status)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected delegated status %d, got %d", http.StatusNotFound, rec.Code)
	}
}
