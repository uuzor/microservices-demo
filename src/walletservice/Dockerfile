FROM golang:1.21-alpine

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download || echo "No go.sum file"

# Copy source code
COPY . .

# Build
RUN go build -o walletservice .

# Expose port
EXPOSE 8085

# Run
CMD ["./walletservice"]
