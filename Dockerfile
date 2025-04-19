# syntax=docker/dockerfile:1

FROM golang:1.22-bullseye AS builder

WORKDIR /app

# Copy only necessary files to build Go binaries
COPY go.mod go.sum ./
RUN go mod download

COPY indexer.go searcher.go main.go ./

# Ensure CGO and FTS5 support for SQLite
ENV CGO_ENABLED=1
RUN go build -tags sqlite_fts5 -o indexer indexer.go
RUN go build -tags sqlite_fts5 -o searcher searcher.go
RUN go build -o main main.go

# --- Minimal runtime image ---
FROM debian:bullseye-slim

RUN apt-get update && \
    apt-get install -y sqlite3 poppler-utils && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/indexer /app/searcher /app/main /app/

# Default volumes for persistent data access
VOLUME ["/app/db", "/app/pdfs"]

ENTRYPOINT ["/bin/bash"]
