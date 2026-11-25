package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"unicode"
)

func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

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

var errInvalidID = errors.New("invalid id format")

func readID(r *http.Request) (string, error) {
	id := r.PathValue("id")
	if id == "" || len(id) > 64 {
		return "", errInvalidID
	}
	for _, char := range id {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '-' {
			return "", errInvalidID
		}
	}
	return id, nil
}

func readString(query url.Values, key, defaultValue string) string {
	if value := query.Get(key); value != "" {
		return value
	}
	return defaultValue
}

func readInt(query url.Values, key string, defaultValue int) (int, error) {
	value := query.Get(key)
	if value == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(value)
}

func readBool(query url.Values, key string) (bool, bool) {
	value := query.Get(key)
	if value == "" {
		return false, false
	}
	parsed, err := strconv.ParseBool(value)
	return parsed, err == nil
}

func readSort(query url.Values, allowed []string, defaultValue string) (string, bool) {
	value := readString(query, "sort", defaultValue)
	return value, slices.Contains(allowed, value)
}
