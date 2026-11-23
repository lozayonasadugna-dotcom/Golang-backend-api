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

func TestCreateOrderNormalisesPhoneAndReturnsFieldErrors(t *testing.T) {
	app := newTestApplication()
	req := httptest.NewRequest(http.MethodPost, "/v1/orders",
		strings.NewReader(`{"area":" Bole ","phone":"+251 911-223344","dish_ids":["d1"]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	app.routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusCreated, rec.Body)
	}
	var response struct {
		Data order.Order `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.Phone != "0911223344" {
		t.Errorf("phone = %q, want %q", response.Data.Phone, "0911223344")
	}

	req = httptest.NewRequest(http.MethodPost, "/orders",
		strings.NewReader(`{"area":"Ayat","phone":"9112233","dish_ids":["d1","d1"]}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	app.routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid request status = %d, want %d; body: %s", rec.Code, http.StatusUnprocessableEntity, rec.Body)
	}
	var invalidResponse struct {
		Error struct {
			Fields map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &invalidResponse); err != nil {
		t.Fatalf("decode validation response: %v", err)
	}
	for _, field := range []string{"area", "phone", "dish_ids"} {
		if invalidResponse.Error.Fields[field] == "" {
			t.Errorf("expected validation error for %q, got %v", field, invalidResponse.Error.Fields)
		}
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/orders",
		strings.NewReader(`{"area":"Bole","phone":"0911223344","dish_ids":["d999"]}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	app.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "dish d999 does not exist") {
		t.Errorf("unknown dish response = %d %s, want a 422 field validation error", rec.Code, rec.Body)
	}
}

func TestListDishesSupportsPaginationAndETags(t *testing.T) {
	app := newTestApplication()
	req := httptest.NewRequest(http.MethodGet, "/v1/dishes?fasting=true&sort=-price&limit=1", nil)
	rec := httptest.NewRecorder()
	app.routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body)
	}
	etag := rec.Header().Get("ETag")
	if etag == "" {
		t.Fatal("expected an ETag response header")
	}
	var response struct {
		Data []menu.Dish `json:"data"`
		Meta struct {
			Total int `json:"total"`
			Limit int `json:"limit"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode menu response: %v", err)
	}
	if len(response.Data) != 1 || response.Data[0].ID != "d7" {
		t.Errorf("first page = %+v, want the highest-priced fasting dish d7", response.Data)
	}
	if response.Meta.Total != 4 || response.Meta.Limit != 1 {
		t.Errorf("meta = %+v, want total 4 and limit 1", response.Meta)
	}

	nextPageReq := httptest.NewRequest(http.MethodGet, "/v1/dishes?fasting=true&sort=-price&limit=1&page=2", nil)
	nextPageRec := httptest.NewRecorder()
	app.routes().ServeHTTP(nextPageRec, nextPageReq)
	if nextPageRec.Header().Get("ETag") == etag {
		t.Error("different menu pages must have different ETags")
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/dishes?fasting=true&sort=-price&limit=1", nil)
	req.Header.Set("If-None-Match", etag)
	rec = httptest.NewRecorder()
	app.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotModified {
		t.Errorf("conditional status = %d, want %d", rec.Code, http.StatusNotModified)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("304 response body = %q, want empty", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/dishes?page=999999999999&limit=100", nil)
	rec = httptest.NewRecorder()
	app.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("out-of-range page status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"data":[]`) {
		t.Errorf("out-of-range page should return an empty array, got %s", rec.Body)
	}
}

func TestOrderStatusRequiresSequentialTransitions(t *testing.T) {
	app := newTestApplication()
	newOrder := order.New("Bole", "0911223344", []menu.Dish{{ID: "d1", PriceSantim: 12000}})
	if _, err := newOrder.ComputeTotal(); err != nil {
		t.Fatalf("compute total: %v", err)
	}
	if err := app.orders.Create(&newOrder); err != nil {
		t.Fatalf("create order: %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, "/v1/orders/"+newOrder.ID,
		strings.NewReader(`{"status":"accepted"}`))
	rec := httptest.NewRecorder()
	app.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("accepted transition status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body)
	}

	req = httptest.NewRequest(http.MethodPatch, "/v1/orders/"+newOrder.ID,
		strings.NewReader(`{"status":"delivered"}`))
	rec = httptest.NewRecorder()
	app.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Errorf("skipped transition status = %d, want %d; body: %s", rec.Code, http.StatusConflict, rec.Body)
	}
}
