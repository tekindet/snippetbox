package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecureHeaders(t *testing.T) {
	rr := httptest.NewRecorder()

	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("OK"))
	})

	secureHeaders(next).ServeHTTP(rr, r)

	rs := rr.Result()

	frameOptions := rs.Header.Get("X-Frame-Options")
	if frameOptions != "deny" {
		t.Errorf("expected %q; got %q", "deny", frameOptions)
	}

	contentTypeOpts := rs.Header.Get("X-Content-Type-Options")
	if contentTypeOpts != "nosniff" {
		t.Errorf("expected %q; got %q", "nosniff", contentTypeOpts)
	}

	referrerPolicy := rs.Header.Get("Referrer-Policy")
	if referrerPolicy != "strict-origin-when-cross-origin" {
		t.Errorf("expected %q; got %q", "strict-origin-when-cross-origin", referrerPolicy)
	}

	csp := rs.Header.Get("Content-Security-Policy")
	if csp == "" {
		t.Error("expected Content-Security-Policy header to be set")
	}

	if rs.StatusCode != http.StatusOK {
		t.Errorf("expected %d; got %d", http.StatusOK, rs.StatusCode)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "OK" {
		t.Errorf("want body to equal %q", "OK")
	}
}
