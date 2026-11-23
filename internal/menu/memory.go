package menu

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type memoryStore struct {
	dishes []Dish
}

// NewMemoryStore creates and returns a menu store.
func NewMemoryStore() *memoryStore {
	return &memoryStore{
		dishes: []Dish{
			{
				ID:          "d1",
				Name:        "Shiro",
				PriceSantim: 12000,
				IsFasting:   true,
				Note:        "Traditional Ethiopian dish",
			},
			{
				ID:          "d2",
				Name:        "Doro Wot",
				PriceSantim: 25000,
				IsFasting:   false,
			},
			{
				ID:          "d3",
				Name:        "Misir Wot",
				PriceSantim: 15000,
				IsFasting:   true,
			},
			{
				ID:          "d4",
				Name:        "Tibs",
				PriceSantim: 28000,
				IsFasting:   false,
			},
			{
				ID:          "d5",
				Name:        "Kitfo",
				PriceSantim: 35000,
				IsFasting:   false,
			},
			{
				ID:          "d6",
				Name:        "Fasting Firfir",
				PriceSantim: 13000,
				IsFasting:   true,
			},
			{
				ID:          "d7",
				Name:        "Injera with Vegetables",
				PriceSantim: 18000,
				IsFasting:   true,
			},
			{
				ID:          "d8",
				Name:        "Pasta",
				PriceSantim: 20000,
				IsFasting:   false,
			},
		},
	}
}

// All returns every dish.
func (s *memoryStore) All() ([]Dish, error) {
	dishes := make([]Dish, len(s.dishes))
	copy(dishes, s.dishes)
	return dishes, nil
}

// ByID searches for a dish using its ID.
func (s *memoryStore) ByID(id string) (Dish, error) {
	for _, dish := range s.dishes {
		if dish.ID == id {
			return dish, nil
		}
	}

	return Dish{}, ErrNotFound
}

// Page filters, sorts, and returns a bounded page of dishes.
func (s *memoryStore) Page(q Query, page, limit int) ([]Dish, int, error) {
	if page < 1 || limit < 1 {
		return nil, 0, errors.New("page and limit must be positive")
	}

	filtered := make([]Dish, 0, len(s.dishes))
	for _, dish := range s.dishes {
		if q.Fasting != nil && dish.IsFasting != *q.Fasting {
			continue
		}
		filtered = append(filtered, dish)
	}

	sort.Slice(filtered, func(i, j int) bool {
		switch q.Sort {
		case "-name":
			return strings.Compare(filtered[i].Name, filtered[j].Name) > 0
		case "price":
			return filtered[i].PriceSantim < filtered[j].PriceSantim
		case "-price":
			return filtered[i].PriceSantim > filtered[j].PriceSantim
		default:
			return strings.Compare(filtered[i].Name, filtered[j].Name) < 0
		}
	})

	total := len(filtered)
	if total == 0 || page > 1+(total-1)/limit {
		return []Dish{}, total, nil
	}
	offset := (page - 1) * limit
	end := min(offset+limit, total)
	return append([]Dish(nil), filtered[offset:end]...), total, nil
}

// ETag returns a weak validator for a specific filtered menu page.
func (s *memoryStore) ETag(q Query, page, limit int) string {
	dishes, total, _ := s.Page(q, page, limit)
	hash := sha256.New()
	fasting := "all"
	if q.Fasting != nil {
		fasting = fmt.Sprintf("%t", *q.Fasting)
	}
	_, _ = fmt.Fprintf(hash, "%d-%d-%d-%s-%s", page, limit, total, q.Sort, fasting)
	for _, dish := range dishes {
		_, _ = fmt.Fprintf(hash, "%q:%q:%d:%t:%q;", dish.ID, dish.Name, dish.PriceSantim, dish.IsFasting, dish.Note)
	}
	return fmt.Sprintf(`W/"%x"`, hash.Sum(nil)[:8])
}
