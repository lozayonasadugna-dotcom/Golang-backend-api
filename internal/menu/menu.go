package menu

import "errors"

// Dish represents one food item on the menu.
type Dish struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PriceSantim int    `json:"price_santim"`
	IsFasting   bool   `json:"is_fasting"`
	Note        string `json:"note,omitempty"`
}

// DishReader describes the operations needed to read dishes.
type DishReader interface {
	All() ([]Dish, error)
	ByID(id string) (Dish, error)
}

// ErrNotFound is returned when a dish does not exist.
var ErrNotFound = errors.New("dish not found")
