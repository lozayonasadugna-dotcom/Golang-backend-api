package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"addis-eats-api/internal/delivery"
	"addis-eats-api/internal/menu"
	"addis-eats-api/internal/order"
)

type createOrderRequest struct {
	Area    string   `json:"area"`
	Phone   string   `json:"phone"`
	DishIDs []string `json:"dish_ids"`
}

// GET /health
func (app *application) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, envelope{
		"data": map[string]string{"status": "ok"},
	})
}

// GET /menu
func (app *application) listDishes(w http.ResponseWriter, r *http.Request) {
	dishes, err := app.dishes.All()
	if err != nil {
		app.serverError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": dishes})
}

// GET /menu/{id}
func (app *application) showDish(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	dish, err := app.dishes.ByID(id)

	switch {
	case errors.Is(err, menu.ErrNotFound):
		app.notFound(w, "dish not found")
		return

	case err != nil:
		app.serverError(w, err)
		return

	default:
		writeJSON(w, http.StatusOK, envelope{"data": dish})
	}
}

// POST /orders
func (app *application) createOrder(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || !strings.EqualFold(mediaType, "application/json") {
		app.badRequest(w, "Content-Type must be application/json")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)

	var input createOrderRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		app.badRequest(w, "malformed JSON body or unknown field")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		app.badRequest(w, "request body must contain one JSON value")
		return
	}

	if len(input.DishIDs) == 0 {
		app.badRequest(w, "dish_ids must not be empty")
		return
	}

	items := make([]menu.Dish, 0, len(input.DishIDs))
	for _, id := range input.DishIDs {
		dish, err := app.dishes.ByID(id)
		if errors.Is(err, menu.ErrNotFound) {
			app.badRequest(w, fmt.Sprintf("unknown dish %s", id))
			return
		}
		if err != nil {
			app.serverError(w, err)
			return
		}
		items = append(items, dish)
	}

	newOrder := order.New(input.Area, input.Phone, items)
	if _, err := newOrder.ComputeTotal(); err != nil {
		switch {
		case errors.Is(err, delivery.ErrUnknownArea):
			app.badRequest(w, fmt.Sprintf("we do not deliver to %s", input.Area))
		case errors.Is(err, order.ErrEmptyOrder):
			app.badRequest(w, err.Error())
		default:
			app.serverError(w, err)
		}
		return
	}

	if err := app.orders.Create(&newOrder); err != nil {
		if errors.Is(err, order.ErrInvalidOrder) {
			app.badRequest(w, "invalid order")
			return
		}
		app.serverError(w, err)
		return
	}

	w.Header().Set("Location", "/orders/"+newOrder.ID)
	writeJSON(w, http.StatusCreated, envelope{"data": newOrder})
}

// DELETE /orders/{id}
func (app *application) cancelOrder(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))

	err := app.orders.Cancel(id)
	if err != nil {
		if errors.Is(err, order.ErrNotFound) {
			app.notFound(w, "order not found")
			return
		}

		app.serverError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, envelope{
		"data": map[string]string{"message": "order cancelled successfully"},
	})
}

func (app *application) listOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := app.orders.All()
	if err != nil {
		app.serverError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": orders})
}
