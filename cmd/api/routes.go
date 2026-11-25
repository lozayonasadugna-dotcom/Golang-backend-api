package main

import "net/http"

type middleware func(http.Handler) http.Handler

func chain(h http.Handler, m ...middleware) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", app.health)
	mux.HandleFunc("GET /v1/health", app.health)

	// Menu routes
	mux.HandleFunc("GET /menu", app.listDishes)
	mux.HandleFunc("GET /menu/{id}", app.showDish)
	mux.HandleFunc("GET /v1/dishes", app.listDishes)
	mux.HandleFunc("GET /v1/dishes/{id}", app.showDish)

	// Order routes
	mux.HandleFunc("GET /orders", app.listOrders)
	mux.HandleFunc("POST /orders", app.createOrder)
	mux.HandleFunc("DELETE /orders/{id}", app.cancelOrder)
	mux.HandleFunc("POST /v1/orders", app.createOrder)
	mux.HandleFunc("GET /v1/orders", app.listOrders)
	mux.HandleFunc("GET /v1/orders/{id}", app.showOrder)
	mux.HandleFunc("PATCH /v1/orders/{id}", app.patchOrder)
	mux.HandleFunc("/", app.notFoundRoute)

	return chain(
		mux,
		app.recoverPanic,
		app.requestID,
		app.logRequest,
		app.secureHeaders,
		app.enableCORS,
	)
}

func (app *application) notFoundRoute(w http.ResponseWriter, r *http.Request) {
	app.notFound(w, "the requested resource was not found")
}
