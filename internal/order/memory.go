package order

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"

	"addis-eats-api/internal/menu"
)

// MemoryStore stores orders in memory.
type MemoryStore struct {
	mu     sync.RWMutex
	orders map[string]Order
}

// NewMemoryStore creates a new order storage.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		orders: make(map[string]Order),
	}
}

func newID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("generate order ID: %w", err)
	}

	return "o-" + hex.EncodeToString(value[:]), nil
}

// Create assigns server-controlled metadata and saves an order.
func (s *MemoryStore) Create(o *Order) error {
	if o == nil || len(o.Items) == 0 || o.TotalSantim <= 0 {
		return ErrInvalidOrder
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for {
		id, err := newID()
		if err != nil {
			return err
		}
		if _, exists := s.orders[id]; exists {
			continue
		}

		o.ID = id
		o.PlacedAt = time.Now().UTC()
		o.Items = append([]menu.Dish(nil), o.Items...)
		s.orders[id] = *o
		return nil
	}
}

// All returns orders in placement order.
func (s *MemoryStore) All() ([]Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]Order, 0, len(s.orders))
	for _, o := range s.orders {
		o.Items = append([]menu.Dish(nil), o.Items...)
		orders = append(orders, o)
	}

	sort.Slice(orders, func(i, j int) bool {
		return orders[i].PlacedAt.Before(orders[j].PlacedAt)
	})

	return orders, nil
}

// Cancel deletes an order using its ID.
func (s *MemoryStore) Cancel(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.orders[id]; !exists {
		return ErrNotFound
	}

	delete(s.orders, id)
	return nil
}
