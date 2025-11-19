FROM golang:1.21-alpine

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download || echo "No go.sum file"

# Copy source code
COPY . .

# Build
RUN go build -o feeservice .

# Expose port
EXPOSE 8087

# Run
CMD ["./feeservice"]
