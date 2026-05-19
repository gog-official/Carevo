# Free Hosting Guide

## The Stack

| Component | Service | Free Tier | Credit Card? |
|-----------|---------|-----------|--------------|
| Go API | HF Spaces Docker / Fly.io | 512 MB RAM | ❌ No (HF) |
| PostgreSQL | Supabase | 500 MB | ❌ No |
| Redis | Upstash | 10 MB / 5K cmd/day | ❌ No |
| AI | Gemini API | 60 req/min | ❌ No |

AI runs on Google's servers — no GPU or RAM needed on your side.

---

## Free Services (No Credit Card)

### PostgreSQL: Supabase

```
✓ 500 MB
✓ No credit card
✓ Sign up with GitHub
✗ Pauses after 1 week idle
```

Get connection string in Settings → Database.

### Redis: Upstash

```
✓ 10 MB
✓ 5,000 commands/day
✓ No credit card
✓ Fine for 5-min TTL cache
```

### AI: Google Gemini API

No credit card. No GPU. Just a Google account:

```bash
# 1. Go to https://aistudio.google.com/apikey
# 2. Click "Create API Key" — instant, free
# 3. Add to .env:
GEMINI_API_KEY=your-key
GEMINI_MODEL=gemini-2.0-flash
```

### Compute (Go API)

| Platform | RAM | Card? | Cold Start | Notes |
|----------|-----|-------|------------|-------|
| **HF Spaces Docker** | 16 GB shared | ❌ **No** | ~10s | Best option, no card ever |
| **Fly.io** | 256 MB × 3 VMs | Maybe | None | Try without card first |
| **Koyeb** | 512 MB | Maybe | None | 1 GB SSD included |

#### HF Spaces (recommended — truly no card)

The Dockerfile has two targets: `api` (serves requests) and `seed` (one-shot data loader). HF Spaces runs the default target, which is `api`. The API auto-runs pending migrations on startup but does **not** seed data.

To get data into the database on HF Spaces:

```bash
# 1. Create Space → Docker → Connect repo
# 2. Set env vars in Space settings (see .env below)
# 3. Deploy — app starts, migrations auto-run, careers table is empty

# 4. Seed the database with a one-off curl:
#    Open Space → "Factory" → "Start a new terminal" then:
cd /app && ./seed

#    Or from your machine, port-forward and run:
docker compose run --rm seed

#    Or seed via API calls if you build a seed endpoint later.
```

The seed container runs migrations first, then inserts 96 careers + 20 survey questions. It's idempotent — safe to re-run.

---

## .env For Free Hosting

```env
DATABASE_URL=postgres://user:pass@db.supabase.co:5432/postgres
REDIS_URL=redis://default:token@upstash.io:6379
GEMINI_API_KEY=your-free-gemini-key
GEMINI_MODEL=gemini-2.0-flash
JWT_SECRET=generate-a-random-secret
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=720h
PORT=8080
```

---

## What Won't Work (No Card)

| Platform | Why |
|----------|-----|
| Vercel / Netlify | No Go runtime |
| Cloudflare Workers | No persistent processes |
| AWS / GCP / Azure | All require credit card |
| Render / Railway | Require card even for free tier |
