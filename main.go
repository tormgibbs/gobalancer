package main

import (
	"flag"
	"log"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"time"
)

type config struct {
	port                int
	healthCheckInterval time.Duration
}

type application struct {
	config   config
	logger   *log.Logger
	servers  []*Server
	strategy Strategy
	mu       sync.RWMutex
}

type Server struct {
	url          *url.URL
	alive        bool
	mux						sync.RWMutex
	reverseProxy *httputil.ReverseProxy
}

func (app *application) addServer(serverURL string) {
	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		log.Fatal(err)
	}

	server := &Server{
		url:          parsedURL,
		alive:        true,
		reverseProxy: httputil.NewSingleHostReverseProxy(parsedURL),
	}


	app.servers = append(app.servers, server)

}


func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "Load Balancer server port")
	flag.DurationVar(&cfg.healthCheckInterval, "healthCheckInterval", 5*time.Second, "Health check interval")

	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

	app := &application{
		config: cfg,
		logger: logger,
	}

	app.strategy = &RoundRobinStrategy{app: app}

	app.addServer("http://localhost:8002/")
	app.addServer("http://localhost:8001/")
	app.addServer("http://localhost:8003/")

	err := app.start()
	if err != nil {
		app.logger.Fatal(err)
	}
}
