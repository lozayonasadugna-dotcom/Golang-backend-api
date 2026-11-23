package main

import (
	"errors"
	"net/http"

	"addis-eats-api/internal/menu"
	"addis-eats-api/internal/order"
	"addis-eats-api/internal/validator"
)

func (app *application) failedValidation(w http.ResponseWriter, fields map[string]string) {
	writeJSON(w, http.StatusUnprocessableEntity, envelope{
		"error": map[string]any{
			"message": "the request was not valid",
			"fields":  fields,
		},
	})
}

func (app *application) conflict(w http.ResponseWriter, message string) {
	writeError(w, http.StatusConflict, message)
}

func (app *application) handleError(w http.ResponseWriter, err error) {
	var validationErr *validator.Error
	switch {
	case errors.As(err, &validationErr):
		app.failedValidation(w, validationErr.Fields)
	case errors.Is(err, menu.ErrNotFound), errors.Is(err, order.ErrNotFound):
		app.notFound(w, "resource not found")
	case errors.Is(err, order.ErrIllegalTransition):
		app.conflict(w, err.Error())
	case errors.Is(err, order.ErrInvalidOrder), errors.Is(err, order.ErrEmptyOrder):
		app.badRequest(w, err.Error())
	default:
		app.serverError(w, err)
	}
}
