package main

type Strategy interface {
	NextServer() *Server
}

type RoundRobinStrategy struct {
	current int
	app     *application
}

func (rr *RoundRobinStrategy) NextServer() *Server {
	if len(rr.app.severs) == 0 {
		return nil
	}

	next := rr.current % len(rr.app.severs)
	rr.current++
	return rr.app.severs[next]
}
