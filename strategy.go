package main

import (
	"sync/atomic"
)

type Strategy interface {
	NextServer() *Server
}

type RoundRobinStrategy struct {
	current int64
	app     *application
}

func (rr *RoundRobinStrategy) NextServer() *Server {
	rr.app.mu.RLock()
	defer rr.app.mu.RUnlock()

	severCount := len(rr.app.servers)

	if severCount == 0 {
		return nil
	}

	for i := 0; i < severCount; i++ {
		next := atomic.AddInt64(&rr.current, 1) % int64(severCount)
		server := rr.app.servers[next]

		if server.alive {
			return server
		}
	}

	return nil
}
