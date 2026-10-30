# syntax=docker/dockerfile:1

# ---- Stage 1: build ----
FROM golang:1.27-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /stress-test ./cmd/stress-test

# ---- Stage 2: runtime ----
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /stress-test /stress-test

USER 65534:65534

ENTRYPOINT ["/stress-test"]
