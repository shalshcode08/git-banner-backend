package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecovery_PanicReturns500(t *testing.T) {
	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	handler := Recovery(panicking)
	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: expected %d, got %d", http.StatusInternalServerError, w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(strings.ToLower(body), "internal server error") {
		t.Errorf("body should contain 'internal server error', got %q", body)
	}
}

func TestRecovery_NoPanicPassesThrough(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusAccepted)
	})

	handler := Recovery(next)
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("Recovery did not call next handler when no panic occurred")
	}
	if w.Code != http.StatusAccepted {
		t.Errorf("status: expected %d, got %d", http.StatusAccepted, w.Code)
	}
}

func TestRecovery_PanicWithError(t *testing.T) {
	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(http.ErrAbortHandler)
	})

	handler := Recovery(panicking)
	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	w := httptest.NewRecorder()

	// should not propagate the panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Recovery middleware let a panic escape: %v", r)
		}
	}()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: expected %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
