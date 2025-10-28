# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . ./
RUN go build -o currency-converter

# Runtime stage
FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/currency-converter ./currency-converter

EXPOSE 8080

CMD ["./currency-converter"]
