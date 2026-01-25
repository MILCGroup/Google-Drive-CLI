# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
ARG VERSION=dev

RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w \
        -X github.com/dl-alexandre/gdrv/internal/cli.version=${VERSION}" \
    -o gdrv ./cmd/gdrv

# Final stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' gdrv
USER gdrv

WORKDIR /home/gdrv

# Copy binary from builder
COPY --from=builder /app/gdrv /usr/local/bin/gdrv

ENTRYPOINT ["gdrv"]
CMD ["--help"]
