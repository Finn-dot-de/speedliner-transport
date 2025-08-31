# syntax=docker/dockerfile:1.7

########## Build ##########
FROM --platform=$BUILDPLATFORM golang:1.24 AS builder
WORKDIR /app

# Falls Minor-Version abweicht, zieht Go die passende Toolchain
ENV GOTOOLCHAIN=auto

# Besserer Cache
COPY go.mod go.sum ./
RUN go mod download

# Rest vom Code
COPY . .

# Output-Ordner anlegen und NUR das Main-Paket bauen
# Wenn dein main in ./cmd/server liegt, ersetze "." durch "./cmd/server"
RUN mkdir -p /out && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/server .

########## Runtime ##########
FROM alpine:3.20
WORKDIR /app

# HTTPS + Healthcheck
RUN apk add --no-cache ca-certificates curl

# Binary + (optional) Frontend
COPY --from=builder /out/server ./server
COPY dist ./dist
COPY .env ./

EXPOSE 8080
ENTRYPOINT ["./server"]
