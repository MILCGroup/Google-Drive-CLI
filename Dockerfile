# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown

RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w \
        -X github.com/milcgroup/gdrv/pkg/version.Version=${VERSION} \
        -X github.com/milcgroup/gdrv/pkg/version.GitCommit=${GIT_COMMIT} \
        -X github.com/milcgroup/gdrv/pkg/version.BuildTime=${BUILD_TIME} \
        -X github.com/milcgroup/gdrv/internal/auth.BundledBuildSource=source" \
    -o gdrv ./cmd/gdrv

# Final stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' gdrv
USER gdrv

WORKDIR /home/gdrv

# Copy binary from builder
COPY --from=builder /app/gdrv /usr/local/bin/gdrv

ENTRYPOINT ["gdrv"]
CMD ["--help"]
