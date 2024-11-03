package main

import (
	"fmt"
	"net/http"
)

func (app *application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server := app.strategy.NextServer()

	if server == nil {
		http.Error(w, "no servers available", http.StatusServiceUnavailable)
		return
	}

	server.reverseProxy.ServeHTTP(w, r)
}

func (app *application) start() error {
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
