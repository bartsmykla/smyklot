# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a \
    -installsuffix cgo \
    -ldflags="-s -w" \
    -o smyklot \
    ./cmd/github-action

# Final stage - scratch base for minimal image
FROM scratch

COPY --from=builder /build/smyklot /smyklot

ENTRYPOINT ["/smyklot"]
