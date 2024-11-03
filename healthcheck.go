package main

import (
	"net/http"
	"net/url"
	"time"
)

func (s *Server) setAlive(alive bool) {
	s.mux.Lock()
	s.alive = alive
	s.mux.Unlock()
}

func (s *Server) isAlive(u *url.URL) bool {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(u.String() + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (app *application) healthCheck() {
	for _, server := range app.servers {
		go func(s *Server) {
			ticker := time.NewTicker(app.config.healthCheckInterval)
			for range ticker.C {
				alive := s.isAlive(s.url)
				s.setAlive(alive)
				if !alive {
					app.logger.Printf("Server %v is down!", s.url)
				}
			}
		}(server)
	}
}