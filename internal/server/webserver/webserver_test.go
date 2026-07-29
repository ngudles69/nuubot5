package webserver

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nuubot/internal/toolkit/logging"
)

// Section 1 - Program Flow

func TestWebServerRoutes(t *testing.T) {
	// Step 1: create WebServer
	var log = logging.Create(io.Discard)
	var actual, err = Create(log, "127.0.0.1:0", io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	var server = httptest.NewServer(actual.server.Handler)
	defer server.Close()

	// Step 2: verify routes
	var cases = []struct {
		path        string
		status      int
		contentType string
		body        string
	}{
		{"/", http.StatusOK, "text/html", "Nuubot Server"},
		{"/health", http.StatusOK, "application/json", `{"status":"ok"}`},
		{"/assets/css/app.css", http.StatusOK, "text/css", "--background"},
		{"/missing", http.StatusNotFound, "text/plain", "404 page not found"},
	}
	for _, current := range cases {
		var response, requestErr = http.Get(server.URL + current.path)
		if requestErr != nil {
			t.Fatal(requestErr)
		}
		var body, readErr = io.ReadAll(response.Body)
		var closeErr = response.Body.Close()
		if readErr != nil || closeErr != nil {
			t.Fatal(readErr, closeErr)
		}
		if response.StatusCode != current.status {
			t.Fatalf(
				"%s status = %d, want %d",
				current.path,
				response.StatusCode,
				current.status,
			)
		}
		if !strings.Contains(response.Header.Get("Content-Type"), current.contentType) {
			t.Fatalf(
				"%s content type = %q, want %q",
				current.path,
				response.Header.Get("Content-Type"),
				current.contentType,
			)
		}
		if !strings.Contains(string(body), current.body) {
			t.Fatalf("%s body = %q, want %q", current.path, body, current.body)
		}
	}
}

func TestWebServerRunStopsWithContext(t *testing.T) {
	// Step 1: create cancelled WebServer
	var logOutput bytes.Buffer
	var log = logging.Create(&logOutput)
	var output bytes.Buffer
	var actual, err = Create(log, "127.0.0.1:0", &output)
	if err != nil {
		t.Fatal(err)
	}
	var caller, cancel = context.WithCancel(context.Background())
	cancel()

	// Step 2: run WebServer
	err = actual.Run(caller)
	if err != nil {
		t.Fatal(err)
	}
	var messages = output.String()
	if !strings.Contains(messages, "WebServer started on http://127.0.0.1:") ||
		!strings.Contains(messages, "WebServer stopped") {
		t.Fatalf("console messages = %q", messages)
	}
	var records = logOutput.String()
	if !strings.Contains(records, "WebServer started on http://127.0.0.1:") ||
		!strings.Contains(records, "WebServer stopped") {
		t.Fatalf("log records = %q", records)
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
