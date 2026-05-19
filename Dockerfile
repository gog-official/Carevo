FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/seed ./cmd/seed

FROM gcr.io/distroless/static-debian12:nonroot AS api

WORKDIR /app

COPY --from=builder /app/server /app/server
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080

ENTRYPOINT ["/app/server"]

FROM gcr.io/distroless/static-debian12:nonroot AS seed

WORKDIR /app

COPY --from=builder /app/seed /app/seed
COPY --from=builder /app/migrations /app/migrations

ENTRYPOINT ["/app/seed"]
