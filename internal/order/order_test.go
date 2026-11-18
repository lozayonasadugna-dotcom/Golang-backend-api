package order

import (
	"errors"
	"testing"

	"addis-eats-api/internal/delivery"
	"addis-eats-api/internal/menu"
)

func TestComputeTotalIncludesServiceAndDeliveryFees(t *testing.T) {
	o := New("Bole", "0911223344", []menu.Dish{
		{ID: "d1", PriceSantim: 12000},
	})

	total, err := o.ComputeTotal()
	if err != nil {
		t.Fatalf("ComputeTotal() error = %v", err)
	}
	if want := 17600; total != want {
		t.Errorf("ComputeTotal() = %d, want %d", total, want)
	}
}

func TestComputeTotalWaivesDeliveryAtThreshold(t *testing.T) {
	o := New("Bole", "0911223344", []menu.Dish{
		{ID: "d1", PriceSantim: 30000},
		{ID: "d2", PriceSantim: 20000},
	})

	total, err := o.ComputeTotal()
	if err != nil {
		t.Fatalf("ComputeTotal() error = %v", err)
	}
	if want := 52500; total != want {
		t.Errorf("ComputeTotal() = %d, want %d", total, want)
	}
}

func TestComputeTotalRejectsUnknownArea(t *testing.T) {
	o := New("Ayat", "0911223344", []menu.Dish{
		{ID: "d1", PriceSantim: 12000},
	})

	_, err := o.ComputeTotal()
	if !errors.Is(err, delivery.ErrUnknownArea) {
		t.Errorf("ComputeTotal() error = %v, want %v", err, delivery.ErrUnknownArea)
	}
}

func TestComputeTotalRejectsEmptyOrder(t *testing.T) {
	o := New("Bole", "0911223344", nil)

	_, err := o.ComputeTotal()
	if !errors.Is(err, ErrEmptyOrder) {
		t.Errorf("ComputeTotal() error = %v, want %v", err, ErrEmptyOrder)
	}
}
