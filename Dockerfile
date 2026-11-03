# Build stage
FROM golang:1.23-alpine AS build_base

# Set the Go environment variables
ENV CGO_ENABLED=1
ENV GO111MODULE=on

# Install dependencies for building the Go application and tzdata for setting the timezone
RUN apk add --no-cache git gcc g++ tzdata

# Set time zone to Vietnam
ENV TZ=Asia/Ho_Chi_Minh

# Set the Current Working Directory inside the container
WORKDIR /src

# Copy go.mod and go.sum to cache dependencies
COPY go.mod .
COPY go.sum .

RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go app
RUN go build -o ./out/app ./cmd/api/main.go

# Final stage - Minimal image
FROM alpine:3.17.0

# Install required certificates and tzdata
RUN apk add --no-cache ca-certificates tzdata

# Set time zone to Vietnam
ENV TZ=Asia/Ho_Chi_Minh

# Set working directory
WORKDIR /app

# Copy the built binary and other necessary files from the build stage
COPY --from=build_base /src/out/app /app/webapi
COPY --from=build_base /src/configs /app/configs

# Ensure the binary is executable
RUN chmod +x /app/webapi

# Expose the correct port
EXPOSE 8088

# Start the application
ENTRYPOINT ["./webapi"]
