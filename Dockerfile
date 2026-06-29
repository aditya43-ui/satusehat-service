# Multi-stage build yang lebih optimal
FROM golang:1.25-alpine AS build

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Cache go mod dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build dengan optimasi
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo \
    -ldflags='-w -s -extldflags "-static"' \
    -o main cmd/api/main.go

# Final stage - distroless untuk keamanan
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy timezone data
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary
COPY --from=build /build/main /app/main

# Use non-root user
USER nonroot:nonroot

EXPOSE 8196

ENTRYPOINT ["/app/main"]