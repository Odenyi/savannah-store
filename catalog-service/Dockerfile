FROM golang:1.22 AS builder
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker caching
COPY go.mod go.sum ./
RUN go mod download

# Now copy the rest of the source
COPY . .

# Build the binary
RUN go build -mod=vendor -o catalog-service

# Use a minimal image for runtime
FROM gcr.io/distroless/base
WORKDIR /app
COPY --from=builder /app/catalog-service .
CMD ["./catalog-service"]
