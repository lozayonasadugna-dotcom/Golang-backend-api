package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
	"sort"
	"strings"

	"addis-eats-api/internal/delivery"
	"addis-eats-api/internal/menu"
	"addis-eats-api/internal/order"
	"addis-eats-api/internal/validator"
)

type createOrderRequest struct {
	Area    string   `json:"area"`
	Phone   string   `json:"phone"`
	DishIDs []string `json:"dish_ids"`
}

func (in *createOrderRequest) Normalise() {
	in.Area = strings.TrimSpace(in.Area)

	phone := strings.NewReplacer(" ", "", "-", "").Replace(in.Phone)
	phone = strings.TrimPrefix(phone, "+251")
	phone = strings.TrimPrefix(phone, "251")
	if len(phone) == 9 && !strings.HasPrefix(phone, "0") {
		phone = "0" + phone
	}
	in.Phone = phone
}

func (in createOrderRequest) Validate(v *validator.Validator) {
	v.Check(in.Phone != "", "phone", "must be provided")
	v.Check(validator.Matches(in.Phone, validator.PhoneRx), "phone", "must be a valid TeleBirr number")
	v.Check(in.Area != "", "area", "must be provided")
	v.Check(validator.In(in.Area, delivery.ServedAreas()...), "area", "must be a served delivery area")
	v.Check(len(in.DishIDs) > 0, "dish_ids", "must not be empty")
	v.Check(len(in.DishIDs) <= 50, "dish_ids", "must not contain more than 50 items")
	v.Check(validator.Unique(in.DishIDs), "dish_ids", "must not contain duplicates")
}

type patchOrderRequest struct {
	Status *order.Status `json:"status"`
}

// GET /health
func (app *application) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, envelope{
		"data": map[string]string{"status": "ok"},
	})
}

// GET /menu
func (app *application) listDishes(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, err := readInt(query, "page", 1)
	if err != nil || page < 1 {
		app.badRequest(w, "page must be a positive integer")
		return
	}
	limit, err := readInt(query, "limit", 20)
	if err != nil || limit < 1 {
		app.badRequest(w, "limit must be a positive integer")
		return
	}
	limit = min(limit, 100)

	fasting, fastingOK := readBool(query, "fasting")
	if query.Has("fasting") && !fastingOK {
		app.badRequest(w, "fasting must be true or false")
		return
	}
	sortBy := readString(query, "sort", "name")
	if !validator.In(sortBy, menu.SortFields...) {
		app.badRequest(w, "unknown sort: "+sortBy)
		return
	}

	q := menu.Query{Sort: sortBy}
	if query.Has("fasting") {
		q.Fasting = &fasting
	}

	var dishes []menu.Dish
	var total int
	var etag string
	if reader, ok := app.dishes.(menu.PageReader); ok {
		dishes, total, err = reader.Page(q, page, limit)
		etag = reader.ETag(q, page, limit)
	} else {
		var all []menu.Dish
		all, err = app.dishes.All()
		if err == nil {
			filtered := make([]menu.Dish, 0, len(all))
			for _, dish := range all {
				if q.Fasting == nil || dish.IsFasting == *q.Fasting {
					filtered = append(filtered, dish)
				}
			}
			sort.Slice(filtered, func(i, j int) bool {
				switch q.Sort {
				case "-name":
					return filtered[i].Name > filtered[j].Name
				case "price":
					return filtered[i].PriceSantim < filtered[j].PriceSantim
				case "-price":
					return filtered[i].PriceSantim > filtered[j].PriceSantim
				default:
					return filtered[i].Name < filtered[j].Name
				}
			})
			total = len(filtered)
			if total == 0 || page > 1+(total-1)/limit {
				dishes = []menu.Dish{}
			} else {
				offset := (page - 1) * limit
				end := min(offset+limit, total)
				dishes = append([]menu.Dish(nil), filtered[offset:end]...)
			}
		}
	}
	if err != nil {
		app.serverError(w, err)
		return
	}
	if etag == "" {
		data, marshalErr := json.Marshal(struct {
			Page   int
			Limit  int
			Total  int
			Query  menu.Query
			Dishes []menu.Dish
		}{page, limit, total, q, dishes})
		if marshalErr != nil {
			app.serverError(w, marshalErr)
			return
		}
		sum := sha256.Sum256(data)
		etag = fmt.Sprintf(`W/"%s"`, hex.EncodeToString(sum[:8]))
	}
	if r.Header.Get("If-None-Match") == etag {
		w.Header().Set("Cache-Control", "public, max-age=60")
		w.Header().Set("ETag", etag)
		w.WriteHeader(http.StatusNotModified)
		return
	}

	pages := int(math.Ceil(float64(total) / float64(limit)))
	if pages == 0 {
		pages = 1
	}
	filters := map[string]any{}
	if q.Fasting != nil {
		filters["fasting"] = *q.Fasting
	}
	w.Header().Set("Cache-Control", "public, max-age=60")
	w.Header().Set("ETag", etag)
	writeJSON(w, http.StatusOK, envelope{
		"data": dishes,
		"meta": map[string]any{
			"page": page, "limit": limit, "total": total, "pages": pages,
			"sort": sortBy, "filters": filters,
		},
	})
}

// GET /menu/{id}
func (app *application) showDish(w http.ResponseWriter, r *http.Request) {
	id, err := readID(r)
	if err != nil {
		app.notFound(w, "dish not found")
		return
	}

	dish, err := app.dishes.ByID(id)

	switch {
	case errors.Is(err, menu.ErrNotFound):
		app.notFound(w, "dish not found")
		return

	case err != nil:
		app.serverError(w, err)
		return

	default:
		w.Header().Set("Cache-Control", "public, max-age=60")
		writeJSON(w, http.StatusOK, envelope{"data": dish})
	}
}

// POST /orders
func (app *application) createOrder(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || !strings.EqualFold(mediaType, "application/json") {
		app.badRequest(w, "Content-Type must be application/json")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)

	var input createOrderRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		app.badRequest(w, "malformed JSON body or unknown field")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		app.badRequest(w, "request body must contain one JSON value")
		return
	}

	input.Normalise()
	v := validator.New()
	input.Validate(v)
	if !v.Valid() {
		app.failedValidation(w, v.Errors)
		return
	}

	items := make([]menu.Dish, 0, len(input.DishIDs))
	for _, id := range input.DishIDs {
		dish, err := app.dishes.ByID(id)
		if errors.Is(err, menu.ErrNotFound) {
			app.failedValidation(w, map[string]string{
				"dish_ids": fmt.Sprintf("dish %s does not exist", id),
			})
			return
		}
		if err != nil {
			app.serverError(w, err)
			return
		}
		items = append(items, dish)
	}

	newOrder := order.New(input.Area, input.Phone, items)
	if _, err := newOrder.ComputeTotal(); err != nil {
		app.handleError(w, err)
		return
	}

	if err := app.orders.Create(&newOrder); err != nil {
		app.handleError(w, err)
		return
	}

	location := "/orders/" + newOrder.ID
	if strings.HasPrefix(r.URL.Path, "/v1/") {
		location = "/v1/orders/" + newOrder.ID
	}
	w.Header().Set("Location", location)
	writeJSON(w, http.StatusCreated, envelope{"data": newOrder})
}

func (app *application) showOrder(w http.ResponseWriter, r *http.Request) {
	id, err := readID(r)
	if err != nil {
		app.notFound(w, "order not found")
		return
	}
	store, ok := app.orders.(order.OrderLifecycleStore)
	if !ok {
		app.serverError(w, errors.New("order store does not support order lookup"))
		return
	}
	o, err := store.ByID(id)
	if err != nil {
		app.handleError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, envelope{"data": o})
}

func (app *application) patchOrder(w http.ResponseWriter, r *http.Request) {
	id, err := readID(r)
	if err != nil {
		app.notFound(w, "order not found")
		return
	}
	store, ok := app.orders.(order.OrderLifecycleStore)
	if !ok {
		app.serverError(w, errors.New("order store does not support status updates"))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)
	var input patchOrderRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		app.badRequest(w, "malformed JSON body or unknown field")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		app.badRequest(w, "request body must contain one JSON value")
		return
	}
	if input.Status == nil {
		app.failedValidation(w, map[string]string{"status": "must be provided"})
		return
	}
	updated, err := store.UpdateStatus(id, *input.Status)
	if err != nil {
		app.handleError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, envelope{"data": updated})
}

// DELETE /orders/{id}
func (app *application) cancelOrder(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))

	err := app.orders.Cancel(id)
	if err != nil {
		if errors.Is(err, order.ErrNotFound) {
			app.notFound(w, "order not found")
			return
		}

		app.serverError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, envelope{
		"data": map[string]string{"message": "order cancelled successfully"},
	})
}

func (app *application) listOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := app.orders.All()
	if err != nil {
		app.serverError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": orders})
}
