# Multi-stage Dockerfile optimized for Render.com and Production
# Build stage
FROM golang:alpine AS build_base

ENV CGO_ENABLED=0
ENV GO111MODULE=on

# Install certificates and tzdata
RUN apk add --no-cache git ca-certificates tzdata

ENV TZ=Asia/Ho_Chi_Minh
WORKDIR /src

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build statically linked binary
COPY . .
RUN go build -ldflags="-s -w" -o ./out/app ./cmd/api/main.go

# Production stage - Ultra-lightweight runner
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata \
    && mkdir -p /app/log

ENV TZ=Asia/Ho_Chi_Minh
WORKDIR /app

# Copy binary from build stage
COPY --from=build_base /src/out/app /app/webapi
RUN chmod +x /app/webapi

# Render injects PORT dynamically (default 8088 local fallback)
EXPOSE 8088

ENTRYPOINT ["/app/webapi"]
