package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecureHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("OK"))
		if err != nil {
			t.Fatal(err)
		}
	})

	secureHeaders(handler).ServeHTTP(recorder, request)
	response := recorder.Result()
	assert.Equal(t, response.StatusCode, http.StatusOK)

	assert.Equal(t, "no-store", response.Header.Get("Cache-Control"))
	assert.Equal(t, "frame-ancestors 'none'", response.Header.Get("Content-Security-Policy"))
	assert.Equal(t, "nosniff", response.Header.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", response.Header.Get("X-Frame-Options"))

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, string(body), "OK")
}
