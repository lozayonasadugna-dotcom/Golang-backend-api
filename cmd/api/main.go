package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"addis-eats-api/internal/menu"
	"addis-eats-api/internal/order"
)

type config struct {
	origins []string
}

// application holds the dependencies shared by handlers.
type application struct {
	config config
	logger *log.Logger
	dishes menu.DishReader
	orders order.OrderStore
}

func corsOrigins() []string {
	raw := os.Getenv("CORS_ORIGINS")
	if raw == "" {
		raw = "http://localhost:3000"
	}

	var origins []string
	for _, origin := range strings.Split(raw, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	app := &application{
		config: config{origins: corsOrigins()},
		logger: logger,
		dishes: menu.NewMemoryStore(),
		orders: order.NewMemoryStore(),
	}

	addr := os.Getenv("API_ADDR")

	if addr == "" {
		addr = ":8080"
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      app.routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	logger.Printf("Starting server on %s", addr)

	log.Fatal(srv.ListenAndServe())
}
