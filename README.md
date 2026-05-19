---
title: Carevo
emoji: 🧭
colorFrom: blue
colorTo: green
sdk: docker
app_port: 8080
pinned: false
---

# Carevo

Go API backend for career discovery and AI-powered career matching, built for Nepal.

## Stack

- **Go** 1.26 + Chi router + sqlx + golang-migrate
- **PostgreSQL** 17 + **Redis** 7
- **JWT** auth (access + refresh tokens, bcrypt)
- **Gemini API** — free AI career matching (no credit card)
- **Testcontainers** for integration tests
- **Docker Compose** — multi-stage distroless ~8 MB image

## Quick Start

```bash
cp .env.example .env   # set GEMINI_API_KEY, JWT_SECRET, etc.
docker compose up --build -d
open http://localhost:8080/swagger/
```

## Features

| Area | Details |
|------|---------|
| **Auth** | Register, login, JWT refresh, profile management |
| **Careers** | 96 Nepal-specific careers with NPR salaries, roadmaps, resources, full-text search |
| **Bookmarks** | Save careers, list saved |
| **Challenges** | 30-day challenges with daily check-ins and streak tracking |
| **Project Ideas** | Community-sourced project ideas per career |
| **AI Layer** | Gemini-powered career matching: survey → scored results → roadmap → career score → chat (SSE) |
| **Cache** | Redis with 5-min TTL, silent fallback to DB |
| **Rate Limiter** | Per-IP sliding window (30/min AI, 10/min chat) |

## API Endpoints

All publicly available routes:
- `GET /health`, `GET /ping`
- `POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`
- `GET /careers`, `GET /careers/{slug}`, `GET /careers/search?q=`
- `GET /categories`
- `GET /careers/{id}/resources`, `GET /careers/{id}/roadmap`
- `GET /careers/{id}/projects`

JWT-protected routes:
- `GET /users/me`, `PUT /users/me`
- `POST /users/me/bookmarks`, `GET /users/me/bookmarks`
- `POST /challenges`, `GET /challenges/active`, `POST /challenges/{id}/checkin`, `GET /challenges/{id}/progress`
- `POST /careers/{id}/projects`

JWT + rate-limited AI routes:
- `GET /ai/survey/questions`, `POST /ai/survey/submit`
- `GET /ai/results/me`, `GET /ai/results/me/roadmap`
- `GET /ai/careers/{id}/score`
- `POST /ai/chat` (SSE stream, 10/min)

## Environment

| Var | Default | Required |
|-----|---------|----------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/carevo?sslmode=disable` | Yes |
| `REDIS_URL` | `localhost:6379` | Yes |
| `JWT_SECRET` | — | Yes |
| `GEMINI_API_KEY` | — | No (AI disabled) |
| `PORT` | `8080` | No |

## Development

```bash
make test      # full test suite (Testcontainers)
make seed      # run seed data
make run       # start API server
make build     # compile binaries
make lint      # golangci-lint
```

## License

MIT
