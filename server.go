package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (app *application) start() error {

	go app.healthCheck()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.port),
		Handler: app,
	}

	app.logger.Printf("starting load balancer on port %d", app.config.port)
	err := server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func (app *application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server := app.strategy.NextServer()

	if server == nil {
		http.Error(w, "no servers available", http.StatusServiceUnavailable)
		return
	}

	var bodyBytes []byte
	var err error
	if r.Body != nil {
		bodyBytes, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "error reading request body", http.StatusInternalServerError)
			return
		}
		r.Body.Close()
	}

	r = app.contextSetRetries(r, 0)
	app.attemptRequest(w, r, server, bodyBytes)
}

func (app *application) attemptRequest(w http.ResponseWriter, r *http.Request, server *Server, bodyBytes []byte) {

	if bodyBytes != nil {
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		r.ContentLength = int64(len(bodyBytes))
	}

	server.reverseProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		app.logger.Printf("Proxy error: %v\n\n", err)
		retries := app.contextGetRetries(r)

		if retries < 3 {
			time.Sleep(10 * time.Millisecond)
			r = app.contextSetRetries(r, retries+1)
			app.attemptRequest(w, r, server, bodyBytes)
			return
		}

		server.alive = false
		app.logger.Printf("Marking server %s as dead after %d failed attempts\n\n", server.url, retries)

		nextServer := app.strategy.NextServer()

		if nextServer != nil {
			r = app.contextSetRetries(r, 0)
			app.logger.Printf("Retrying request with different server %s", nextServer.url)

			app.attemptRequest(w, r, nextServer, bodyBytes)
			return
		}
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
	}

	server.reverseProxy.ServeHTTP(w, r)
}
