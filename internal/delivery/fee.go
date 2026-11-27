package delivery

import "errors"

var ErrUnknownArea = errors.New("unknown delivery area")

// Sample delivery fees in santim.

var deliveryFees = map[string]int{
	"Bole":      5000,
	"Piassa":    4000,
	"Mexico":    4500,
	"CMC":       6000,
	"Arat Kilo": 4000,
	"Kazanchis": 3500,
	"Megenagna": 4500,
}

// ServedAreas returns the supported delivery areas.
func ServedAreas() []string {
	areas := make([]string, 0, len(deliveryFees))
	for area := range deliveryFees {
		areas = append(areas, area)
	}
	return areas
}

// FeeByArea returns the delivery fee for an area.
func FeeByArea(area string) (int, error) {
	fee, exists := deliveryFees[area]

	if !exists {
		return 0, ErrUnknownArea
	}

	return fee, nil
}
