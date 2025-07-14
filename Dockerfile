FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git (needed for go mod download)
RUN apk add --no-cache git ca-certificates

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations for smaller binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/

# Copy the binary
COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]
