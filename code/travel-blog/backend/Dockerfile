# syntax=docker/dockerfile:1

FROM golang:1.21-alpine AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o travel-blog ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/travel-blog ./travel-blog
EXPOSE 8080
ENTRYPOINT ["/app/travel-blog"]
