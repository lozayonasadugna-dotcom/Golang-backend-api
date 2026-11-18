package main

import (
	"encoding/json"
	"net/http"
)

type envelope map[string]any

// writeJSON sends data as a JSON response.
func writeJSON(w http.ResponseWriter, status int, data envelope) {
	b, err := json.Marshal(data)

	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(b)
}

// writeError sends an error using a consistent JSON format.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, envelope{
		"error": map[string]string{"message": msg},
	})
}

func (app *application) serverError(w http.ResponseWriter, err error) {
	app.logger.Println("SERVER ERROR:", err)
	writeError(w, http.StatusInternalServerError, "internal error")
}

func (app *application) notFound(w http.ResponseWriter, message string) {
	writeError(w, http.StatusNotFound, message)
}

func (app *application) badRequest(w http.ResponseWriter, message string) {
	writeError(w, http.StatusBadRequest, message)
}
