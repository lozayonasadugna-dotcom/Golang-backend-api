package order

import (
	"errors"
	"strings"
	"time"

	"addis-eats-api/internal/delivery"
	"addis-eats-api/internal/menu"
)

var (
	ErrNotFound     = errors.New("order not found")
	ErrInvalidOrder = errors.New("invalid order")
	ErrEmptyOrder   = errors.New("order must contain at least one dish")
)

// Order represents one customer's order.
type Order struct {
	ID          string      `json:"id"`
	Area        string      `json:"area"`
	Phone       string      `json:"phone"`
	Items       []menu.Dish `json:"items"`
	Status      string      `json:"status"`
	TotalSantim int         `json:"total_santim"`
	PlacedAt    time.Time   `json:"placed_at"`
}

// New creates an order from menu items selected by the customer.
func New(area, phone string, items []menu.Dish) Order {
	return Order{
		Area:   strings.TrimSpace(area),
		Phone:  strings.TrimSpace(phone),
		Items:  append([]menu.Dish(nil), items...),
		Status: "placed",
	}
}

// ComputeTotal calculates the subtotal, service fee, and delivery fee.
func (o *Order) ComputeTotal() (int, error) {
	if len(o.Items) == 0 {
		return 0, ErrEmptyOrder
	}

	subtotal := 0
	for _, item := range o.Items {
		subtotal += item.PriceSantim
	}

	deliveryFee, err := delivery.FeeByArea(o.Area)
	if err != nil {
		return 0, err
	}
	if subtotal >= 50000 {
		deliveryFee = 0
	}

	serviceFee := subtotal / 20
	o.TotalSantim = subtotal + serviceFee + deliveryFee

	return o.TotalSantim, nil
}

// OrderStore defines the operations for managing orders.
type OrderStore interface {
	Create(o *Order) error
	All() ([]Order, error)
	Cancel(id string) error
}

// OrderLifecycleStore supports reading and advancing an individual order.
type OrderLifecycleStore interface {
	ByID(id string) (Order, error)
	UpdateStatus(id string, status Status) (Order, error)
}
