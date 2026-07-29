// Package webserver owns the Nuubot Server HTTP boundary.
package webserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"time"

	"nuubot/internal/toolkit/logging"
	"nuubot/web"
)

const (
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 5 * time.Second
)

// WebServer owns one HTTP service.
type WebServer struct {
	log      *logging.Logger
	output   io.Writer
	server   *http.Server
	homePage *template.Template
}

// Section 1 - Program Flow

// Create constructs one stopped WebServer.
func Create(
	log *logging.Logger,
	address string,
	output io.Writer,
) (*WebServer, error) {
	// Step 1: validate WebServer inputs
	if log == nil || address == "" || output == nil {
		return nil, fmt.Errorf("create webserver requires log, address, and output")
	}

	// Step 2: parse home template
	var homePage, err = template.ParseFS(web.Files, "templates/index.html")
	if err != nil {
		return nil, fmt.Errorf("create webserver: parse home template: %w", err)
	}

	// Step 3: create WebServer
	var result = &WebServer{
		log:      log,
		output:   output,
		homePage: homePage,
	}
	result.server = &http.Server{
		Addr:              address,
		Handler:           result.routes(),
		ReadHeaderTimeout: readHeaderTimeout,
	}
	return result, nil
}

// Run serves HTTP until cancellation or service failure.
func (s *WebServer) Run(caller context.Context) error {
	// Step 1: open HTTP listener
	if caller == nil {
		return fmt.Errorf("run webserver requires context")
	}
	var listener, err = net.Listen("tcp", s.server.Addr)
	if err != nil {
		return fmt.Errorf("run webserver: listen on %s: %w", s.server.Addr, err)
	}

	// Step 2: start HTTP server
	var serveErrors = make(chan error, 1)
	go func() {
		serveErrors <- s.server.Serve(listener)
	}()
	var startedMessage = fmt.Sprintf(
		"WebServer started on http://%s",
		listener.Addr(),
	)
	s.log.Info(startedMessage)
	s.printConsole(startedMessage)

	// Step 3: wait for WebServer stop
	select {
	case err = <-serveErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("run webserver: serve HTTP: %w", err)
	case <-caller.Done():
	}

	// Step 4: stop HTTP server
	var shutdownContext, cancel = context.WithTimeout(
		context.Background(),
		shutdownTimeout,
	)
	var shutdownErr = s.server.Shutdown(shutdownContext)
	cancel()
	if shutdownErr != nil {
		var closeErr = s.server.Close()
		shutdownErr = errors.Join(shutdownErr, closeErr)
	}
	var serveErr = <-serveErrors
	if errors.Is(serveErr, http.ErrServerClosed) {
		serveErr = nil
	}
	if shutdownErr != nil || serveErr != nil {
		return fmt.Errorf(
			"run webserver: stop HTTP: %w",
			errors.Join(shutdownErr, serveErr),
		)
	}

	// Step 5: log run completed
	const stoppedMessage = "WebServer stopped"
	s.log.Info(stoppedMessage)
	s.printConsole(stoppedMessage)
	return nil
}

// Section 2 - Domain Helpers

func (s *WebServer) routes() http.Handler {
	var routes = http.NewServeMux()
	routes.HandleFunc("GET /{$}", s.home)
	routes.HandleFunc("GET /health", s.health)
	routes.Handle("GET /assets/", http.FileServerFS(web.Files))
	return routes
}

func (s *WebServer) home(response http.ResponseWriter, _ *http.Request) {
	var output bytes.Buffer
	var err = s.homePage.Execute(&output, nil)
	if err != nil {
		s.log.Error(fmt.Sprintf("render home page failed: %v", err))
		http.Error(response, "internal server error", http.StatusInternalServerError)
		return
	}

	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = response.Write(output.Bytes())
	if err != nil {
		s.log.Error(fmt.Sprintf("write home page failed: %v", err))
	}
}

func (s *WebServer) health(response http.ResponseWriter, _ *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var _, err = io.WriteString(response, `{"status":"ok"}`+"\n")
	if err != nil {
		s.log.Error(fmt.Sprintf("write health response failed: %v", err))
	}
}

// Section 3 - Generic Helpers

func (s *WebServer) printConsole(message string) {
	var _, err = fmt.Fprintln(s.output, message)
	if err != nil {
		s.log.Warning(fmt.Sprintf("write WebServer console message failed: %v", err))
	}
}
