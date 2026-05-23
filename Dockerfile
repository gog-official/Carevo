FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/seed ./cmd/seed

FROM alpine:3.21

RUN adduser -D -u 1000 user && \
    apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder --chown=user /app/server /app/server
COPY --from=builder --chown=user /app/seed /app/seed
COPY --from=builder --chown=user /app/migrations /app/migrations

USER user

EXPOSE 8080

COPY --chown=user <<'EOF' /app/entrypoint.sh
#!/bin/sh
if [ "$SERVICE_TYPE" = "seed" ]; then
  exec /app/seed
fi
exec /app/server
EOF

RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]
