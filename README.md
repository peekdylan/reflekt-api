# Reflekt API

A RESTful backend API for Reflekt, an AI-powered journaling app that analyzes your mood and provides insights using Claude AI.

Built with Go, PostgreSQL, and the Anthropic Claude API.

---

## Features

- User registration and login with JWT authentication
- Create, read, and delete journal entries
- Automatic AI mood analysis powered by Claude — runs in the background after every entry
- Secure password hashing with bcrypt
- PostgreSQL database with migration support via Goose
- SQLC-generated type-safe database queries
- CORS support for local frontend development
- Docker support for the database

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go |
| Database | PostgreSQL |
| Migrations | Goose |
| Query Generation | SQLC |
| Authentication | JWT (golang-jwt) |
| Password Hashing | bcrypt |
| AI Analysis | Anthropic Claude API |
| Containerization | Docker |

---

## Project Structure

    reflekt-api/
    ├── cmd/
    │   └── api/
    │       └── main.go          # Entry point — sets up DB, config, and routes
    ├── internal/
    │   ├── ai/
    │   │   └── analyzer.go      # Claude AI integration for mood analysis
    │   ├── api/
    │   │   ├── auth.go          # Register and login handlers
    │   │   ├── config.go        # Shared API config and dependencies
    │   │   ├── entries.go       # Journal entry CRUD handlers
    │   │   ├── health.go        # Health check endpoint
    │   │   ├── json.go          # JSON response helpers
    │   │   └── middleware.go    # JWT auth and CORS middleware
    │   └── database/
    │       ├── helpers.go       # Database utility functions
    │       └── ...              # SQLC generated files
    ├── sql/
    │   ├── queries/             # Raw SQL queries used by SQLC
    │   └── schema/              # Goose migration files
    ├── .env.example             # Environment variable template
    ├── docker-compose.yml       # PostgreSQL container setup
    ├── sqlc.yaml                # SQLC configuration
    └── go.mod                   # Go module dependencies

---

## Getting Started

### Prerequisites

- Go 1.22+
- Docker
- Goose (`go install github.com/pressly/goose/v3/cmd/goose@latest`)
- SQLC (`brew install sqlc`)
- An Anthropic API key ([console.anthropic.com](https://console.anthropic.com))

### Setup

1. Clone the repository

        git clone https://github.com/peekdylan/reflekt-api
        cd reflekt-api

2. Install dependencies

        go mod download

3. Copy the environment variable template and fill in your values

        cp .env.example .env

4. Start the PostgreSQL database with Docker

        docker compose up -d

5. Run database migrations

        goose -dir sql/schema postgres "postgres://reflekt:reflektpass@127.0.0.1:5432/reflekt?sslmode=disable" up

6. Start the API server

        go run cmd/api/main.go

The API will be running at `http://localhost:8080`

---

## API Endpoints

### Auth

| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| POST | `/v1/register` | Create a new account | No |
| POST | `/v1/login` | Login and receive JWT token | No |

### Entries

| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| GET | `/v1/entries` | Get all entries for the logged-in user | Yes |
| POST | `/v1/entries` | Create a new journal entry | Yes |
| DELETE | `/v1/entries/{id}` | Delete an entry by ID | Yes |

### Health

| Method | Endpoint | Description |
|---|---|---|
| GET | `/v1/health` | Check if the API is running |

---

## Environment Variables

See `.env.example` for all required variables:

    PORT=8080
    DB_URL=postgres://reflekt:reflektpass@127.0.0.1:5432/reflekt?sslmode=disable
    JWT_SECRET=your-secret-key
    ANTHROPIC_API_KEY=your-anthropic-api-key

---

## How AI Analysis Works

When a user creates a journal entry the API responds immediately with the saved entry. In the background a goroutine sends the entry title and body to the Claude API which analyzes the emotional tone and returns a mood label and a short insight paragraph. These are then saved back to the entry in the database. When the user fetches their entries the mood and analysis are included in the response.

---

## Related

- [Reflekt App](https://github.com/peekdylan/reflekt-app) — The React Native frontend

---

## License

MIT