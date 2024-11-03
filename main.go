package main

import (
	"flag"
	"log"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

type config struct {
	port                int
	healthCheckInterval time.Duration
	// severs              []string
}

type application struct {
	config   config
	logger   *log.Logger
	severs   []*Server
	strategy Strategy
}

type Server struct {
	url          *url.URL
	alive        bool
	reverseProxy *httputil.ReverseProxy
}

func (app *application) addServer(serverURL string) {
	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		log.Fatal(err)
	}

	server := &Server{
		url: parsedURL,
		alive: true,
		reverseProxy: httputil.NewSingleHostReverseProxy(parsedURL),
	}

	app.severs = append(app.severs, server)

}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "Load Balancer server port")
	flag.DurationVar(&cfg.healthCheckInterval, "healthCheckInterval", 5*time.Second, "Health check interval")

	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	app := &application{
		config: cfg,
		logger: logger,
	}

	app.strategy = &RoundRobinStrategy{app: app}

	app.addServer("http://localhost:8001/")
	app.addServer("http://localhost:8002/")
	app.addServer("http://localhost:8003/")

	err := app.start()
	if err != nil {
		app.logger.Fatal(err)
	}
}
