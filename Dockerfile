# ---------- Build stage ----------
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /scraper ./cmd/scraper

# ---------- Runtime stage ----------
FROM chromedp/headless-shell:latest

COPY --from=builder /scraper /usr/local/bin/scraper

# Default output directory inside container
RUN mkdir -p /data/outputs

ENTRYPOINT ["scraper"]
CMD ["--help"]
