package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"addis-eats-api/internal/menu"
	"addis-eats-api/internal/order"
)

// application holds the dependencies shared by handlers.
type application struct {
	logger *log.Logger
	dishes menu.DishReader
	orders order.OrderStore
}

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	app := &application{
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
