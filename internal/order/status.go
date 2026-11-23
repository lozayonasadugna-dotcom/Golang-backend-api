package order

import (
	"encoding/json"
	"errors"
	"fmt"
)

var ErrIllegalTransition = errors.New("illegal status transition")

// Status represents a step in the order preparation and delivery workflow.
type Status int

const (
	StatusPlaced Status = iota
	StatusAccepted
	StatusCooking
	StatusOnTheWay
	StatusDelivered
)

func (s Status) String() string {
	switch s {
	case StatusPlaced:
		return "placed"
	case StatusAccepted:
		return "accepted"
	case StatusCooking:
		return "cooking"
	case StatusOnTheWay:
		return "on_the_way"
	case StatusDelivered:
		return "delivered"
	default:
		return "unknown"
	}
}

func ParseStatus(value string) (Status, error) {
	switch value {
	case "placed":
		return StatusPlaced, nil
	case "accepted":
		return StatusAccepted, nil
	case "cooking":
		return StatusCooking, nil
	case "on_the_way":
		return StatusOnTheWay, nil
	case "delivered":
		return StatusDelivered, nil
	default:
		return Status(-1), fmt.Errorf("invalid status %q", value)
	}
}

func (s Status) CanTransitionTo(next Status) bool {
	return s >= StatusPlaced && s < StatusDelivered && next == s+1
}

func (s Status) MarshalJSON() ([]byte, error) {
	if s < StatusPlaced || s > StatusDelivered {
		return nil, fmt.Errorf("invalid order status %d", s)
	}
	return json.Marshal(s.String())
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	status, err := ParseStatus(value)
	if err != nil {
		return err
	}
	*s = status
	return nil
}
