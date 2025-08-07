# Basis: Minimales Ubuntu-Image
FROM ubuntu:24.04

# Arbeitsverzeichnis im Container
WORKDIR /app

# Benötigte Tools installieren (z. B. SSL, DNS-Zertifikate)
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY server ./
COPY .env ./

EXPOSE 8080

# Start-Befehl
CMD ["./server"]
