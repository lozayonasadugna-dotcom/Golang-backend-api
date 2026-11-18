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
│   └── order/
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

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| GET | `/health` | Check if the server is running |
| GET | `/menu` | View all dishes |
| GET | `/menu/{id}` | View one dish |
| POST | `/orders` | Create an order |
| GET | `/orders` | View all orders |
| DELETE | `/orders/{id}` | Cancel an order |

## Running Tests

To run the project tests, use:

```bash
go test ./...
```

## Note

Orders are stored in memory, so they will be cleared when the server is restarted.