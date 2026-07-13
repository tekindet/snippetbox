package main

import (
	"bytes"
	"net/http"
	"testing"
)

func TestPing(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/ping")

	if code != http.StatusOK {
		t.Errorf("expected %q; got %q", http.StatusOK, code)
	}

	if string(body) != "OK" {
		t.Errorf("expected body to contain %v", "OK")
	}
}

func TestShowSnippet(t *testing.T) {
	// Create an instance of the application using `netTestApplication() function`
	app := newTestApplication(t)

	// Create a server
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody []byte
	}{
		{"Valid ID", "/snippets/1", http.StatusOK, []byte("I never meant for any of this to happen")},
		{"Non existent ID", "/snippets/2", http.StatusNotFound, nil},
		{"Negative ID", "/snippets/-10", http.StatusNotFound, nil},
		{"Empty ID", "/snippets/", http.StatusNotFound, nil},
		{"Trailing Slash", "/snippets/1/", http.StatusNotFound, nil},
		{"String ID", "/snippets/foo", http.StatusNotFound, nil},
		{"Decimal ID", "/snippets/1.3", http.StatusNotFound, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)
			if code != tt.wantCode {
				t.Errorf("expected %d; got %d", tt.wantCode, code)
			}

			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("expected body to contain %q", tt.wantBody)
			}
		})
	}
}
