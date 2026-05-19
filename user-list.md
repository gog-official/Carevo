# Setup Checklist

## Quick Start (Docker Compose)

```bash
# 1. Clone and go
git clone <repo> && cd carevo

# 2. Set your Gemini API key (free, no credit card)
# Get one at https://aistudio.google.com/apikey
export GEMINI_API_KEY=your-key-here

# 3. Start everything
docker compose up --build -d

# 4. Open
open http://localhost:8080/swagger/
```

**What happens:**
- PostgreSQL 17 boots
- Redis 7 boots
- `seed` container runs migrations + seeds 96 Nepal careers + 20 survey questions → exits
- `api` container starts, auto-runs any pending migrations, serves on `:8080`

**No manual migration or seed steps.** Everything auto-runs.

---

## Local Dev (No Docker)

```bash
cp .env.example .env
# Edit .env with your local PostgreSQL/Redis

# Run migrations + seed
make seed

# Start API
make run

# Tests
make test
```

---

## What's In the Box

### Auth
| Endpoint | Description |
|---|---|
| `POST /auth/register` | Create account |
| `POST /auth/login` | Get JWT tokens |
| `POST /auth/refresh` | Rotate refresh token |
| `GET /users/me` | Profile |
| `PUT /users/me` | Update profile |

### Careers (96 total, Nepal-specific)
| Endpoint | Description |
|---|---|
| `GET /careers` | Paginated list, filter by category/tag |
| `GET /careers/{slug}` | Single career detail |
| `GET /careers/search?q=` | Full-text search |
| `GET /careers/{id}/resources` | Learning resources |
| `GET /careers/{id}/roadmap` | Step-by-step roadmap with external links |
| `GET /categories` | Categories with career counts |

### Bookmarks
| Endpoint | Auth |
|---|---|
| `POST /users/me/bookmarks` | JWT |
| `GET /users/me/bookmarks` | JWT |

### 30-Day Challenges
| Endpoint | Auth |
|---|---|
| `POST /challenges` | JWT |
| `GET /challenges/active` | JWT |
| `POST /challenges/{id}/checkin` | JWT |
| `GET /challenges/{id}/progress` | JWT |

### Project Ideas
| Endpoint | Auth |
|---|---|
| `POST /careers/{id}/projects` | JWT |
| `GET /careers/{id}/projects` | Open |

### AI Layer (Gemini API — free, no credit card)
| Endpoint | Auth | Rate Limit |
|---|---|---|
| `GET /ai/survey/questions` | JWT | 30/min |
| `POST /ai/survey/submit` | JWT | 30/min |
| `GET /ai/results/me` | JWT | 30/min |
| `GET /ai/results/me/roadmap` | JWT | 30/min |
| `GET /ai/careers/{id}/score` | JWT | 30/min |
| `POST /ai/chat` | JWT (SSE stream) | 10/min |

### Infrastructure
| Feature | Details |
|---|---|
| Redis cache | 5-min TTL on `GET /careers` and `GET /careers/{slug}`, silent fallback |
| Swagger | Auto-generated docs at `/swagger/` |
| Rate limiter | Per-IP sliding window |
| Auto-migrations | `golang-migrate` runs on API startup |
| Docker | Multi-stage, distroless ~8 MB image |

---

## Environment Variables

| Var | Default | Required |
|---|---|---|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/carevo?sslmode=disable` | Yes |
| `REDIS_URL` | `localhost:6379` | Yes |
| `JWT_SECRET` | — | Yes |
| `JWT_ACCESS_TTL` | `15m` | No |
| `JWT_REFRESH_TTL` | `720h` | No |
| `PORT` | `8080` | No |
| `GEMINI_API_KEY` | — | No (AI disabled if empty) |
| `GEMINI_MODEL` | `gemini-2.0-flash` | No |

---

## Commands

```bash
make up        # docker compose up --build -d
make down      # docker compose down
make logs      # docker compose logs -f
make test      # go test ./... -v -count=1 -timeout=300s (spins up real PG via Testcontainers)
make seed      # go run ./cmd/seed
make run       # go run ./cmd/api
make swagger   # regen OpenAPI docs
make build     # compile binaries
make lint      # golangci-lint run
make clean     # rm -rf bin/ + docker compose down -v
make rebuild   # docker compose build --no-cache api && up

./test.sh      # Full CI-style script: vet → build → test → docker build → optional curl
```
