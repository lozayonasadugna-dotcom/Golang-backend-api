# Addis Eats API

## About the Project

Addis Eats API is a simple food ordering REST API built using Go. It allows users to view available food items, place orders, check their orders, and cancel them.

I built this project to practice working with Go, handling HTTP requests, and organizing code into different packages.

## Features

- View all available dishes.
- View a dish by its ID.
- Place a food order.
- Automatically calculate order costs and delivery fees.
- View all orders.
- Cancel an order.
- Handle invalid requests and errors.

## Technologies Used

- Go 1.24.4
- REST API
- JSON
- Go standard library

## Project Structure

```text
addis-eats-api/
├── cmd/
│   └── api/
├── internal/
│   ├── delivery/
│   ├── menu/
│   ├── order/
│   └── validator/
├── go.mod
└── README.md
```

## How to Run

Make sure Go is installed, then open the project folder in your terminal.

Run the following command to start the server:

```bash
go run ./cmd/api
```

The API will run at:

```text
http://localhost:8080
```

The API allows browser requests from `http://localhost:3000` by default. Set
`CORS_ORIGINS` to a comma-separated list to allow other origins, for example:

```powershell
$env:CORS_ORIGINS = "http://localhost:3000,https://example.com"
go run ./cmd/api
```

Responses include an `X-Request-Id` header. You can provide this header on a
request to use your own correlation ID; requests are logged with their ID,
method, path, status, and duration.

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| GET | `/health` | Check if the server is running |
| GET | `/menu` | View all dishes |
| GET | `/menu/{id}` | View one dish |
| POST | `/orders` | Create an order |
| GET | `/orders` | View all orders |
| DELETE | `/orders/{id}` | Cancel an order |

The versioned API aliases are also available: `GET /v1/health`, `GET /v1/dishes`,
`GET /v1/dishes/{id}`, `GET /v1/orders`, `POST /v1/orders`,
`GET /v1/orders/{id}`, and `PATCH /v1/orders/{id}`. Existing routes remain supported.

### Menu pagination and filtering

`GET /menu` (or `GET /v1/dishes`) accepts `page`, `limit`, `fasting`, and `sort`
query parameters. For example, `/v1/dishes?page=1&limit=5&fasting=true&sort=-price`
returns a page of fasting dishes sorted by price, highest first. Limits are capped
at 100. The response includes pagination metadata, and the menu supports ETags and
`If-None-Match` conditional requests.

### Order validation and status

Order creation validates the delivery area, TeleBirr mobile number, dish IDs,
duplicate dishes, and a maximum of 50 dishes. Ethiopian numbers using the `+251`
prefix and spaces or hyphens are normalized to the local `09...` or `07...` format.
Invalid fields return HTTP 422 with a field-error map.

Orders start with status `placed`. Update them one step at a time with
`PATCH /v1/orders/{id}` and a JSON body such as `{"status":"accepted"}`. The
supported progression is `placed` -> `accepted` -> `cooking` -> `on_the_way` ->
`delivered`; skipped or backward transitions return HTTP 409.

## Running Tests

To run the project tests, use:

```bash
go test ./...
```

## Note

Orders are stored in memory, so they will be cleared when the server is restarted.