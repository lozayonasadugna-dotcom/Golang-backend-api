package delivery

import "errors"

var ErrUnknownArea = errors.New("unknown delivery area")

// Sample delivery fees in santim.
// Replace these with the rates from your previous lesson.
var deliveryFees = map[string]int{
	"Bole":      5000,
	"Piassa":    4000,
	"Mexico":    4500,
	"CMC":       6000,
	"Arat Kilo": 4000,
}

// FeeByArea returns the delivery fee for an area.
func FeeByArea(area string) (int, error) {
	fee, exists := deliveryFees[area]

	if !exists {
		return 0, ErrUnknownArea
	}

	return fee, nil
}
