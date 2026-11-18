package menu

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
	return s.dishes, nil
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
