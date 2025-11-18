FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY *.go ./
COPY data/ ./data/

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o marketservice

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/marketservice .
COPY --from=builder /app/data ./data/

EXPOSE 3550

CMD ["./marketservice"]
