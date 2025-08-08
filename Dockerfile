# Build-Stage
FROM golang:1.24.5 as builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o server
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server

# Final image
FROM ubuntu:24.04

WORKDIR /app
COPY --from=builder /app/server ./
COPY .env ./
COPY dist ./dist

RUN apt-get update && \
    apt-get install -y ca-certificates curl && \
    chmod +x ./server && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 8080
CMD ["./server"]

