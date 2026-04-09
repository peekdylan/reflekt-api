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
