package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"addis-eats-api/internal/menu"
	"addis-eats-api/internal/order"
)

func newTestApplication() *application {
	return &application{
		logger: log.New(io.Discard, "", 0),
		dishes: menu.NewMemoryStore(),
		orders: order.NewMemoryStore(),
	}
}

func TestCreateOrderCalculatesPriceFromMenu(t *testing.T) {
	app := newTestApplication()
	body := `{"area":"Bole","phone":"0911223344","dish_ids":["d1"]}`
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	app.createOrder(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusCreated, rec.Body)
	}

	var response struct {
		Data order.Order `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.TotalSantim != 17600 {
		t.Errorf("total_santim = %d, want 17600", response.Data.TotalSantim)
	}
	if response.Data.ID == "" {
		t.Error("expected server-generated order ID")
	}
	if response.Data.Status != "placed" {
		t.Errorf("status = %q, want %q", response.Data.Status, "placed")
	}
	if rec.Header().Get("Location") == "" {
		t.Error("expected Location header")
	}

	listRec := httptest.NewRecorder()
	app.listOrders(listRec, httptest.NewRequest(http.MethodGet, "/orders", nil))
	if listRec.Code != http.StatusOK {
		t.Fatalf("GET /orders status = %d, want %d", listRec.Code, http.StatusOK)
	}

	var listResponse struct {
		Data []order.Order `json:"data"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listResponse); err != nil {
		t.Fatalf("decode order list: %v", err)
	}
	if len(listResponse.Data) != 1 {
		t.Errorf("GET /orders returned %d orders, want 1", len(listResponse.Data))
	}
}

func TestCreateOrderRejectsClientSuppliedTotal(t *testing.T) {
	app := newTestApplication()
	body := `{"area":"Bole","phone":"0911223344","dish_ids":["d1"],"total_santim":1}`
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	app.createOrder(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
