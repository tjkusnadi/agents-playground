# syntax=docker/dockerfile:1

FROM golang:1.22 AS builder
WORKDIR /app

COPY backend/go.* ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server

FROM gcr.io/distroless/base-debian12
WORKDIR /app

COPY --from=builder /app/server ./server
COPY frontend ./frontend

ENV PORT=8080
ENV FRONTEND_DIR=/app/frontend
EXPOSE 8080

USER nonroot:nonroot
CMD ["./server"]
