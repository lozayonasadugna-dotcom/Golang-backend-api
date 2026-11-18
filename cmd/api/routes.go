package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", app.health)

	// Menu routes
	mux.HandleFunc("GET /menu", app.listDishes)
	mux.HandleFunc("GET /menu/{id}", app.showDish)

	// Order routes
	mux.HandleFunc("GET /orders", app.listOrders)
	mux.HandleFunc("POST /orders", app.createOrder)
	mux.HandleFunc("DELETE /orders/{id}", app.cancelOrder)

	return mux
}
