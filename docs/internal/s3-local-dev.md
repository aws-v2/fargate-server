
---
title: Local Development Setup
description: Step-by-step guide to running the S3 service locally including dependencies like MinIO, NATS, PostgreSQL, and environment configuration
icon: code
tags: [development, local, setup, dev-environment, onboarding]
---

 

## Local Development Setup

### Overview

This guide explains how to run the S3 service locally for development, testing, and debugging.

The system depends on multiple infrastructure components:

- MinIO (object storage backend)
- PostgreSQL (metadata storage)
- NATS (event bus)
- S3 service (Go backend)

---

## Prerequisites

Ensure you have installed:

- Go (1.21+)
- Docker & Docker Compose
- Make (optional but recommended)
- Git

---

## Required Services

The system requires the following services running locally:

### 1. MinIO (Object Storage)

Used as the primary storage backend.

```bash
docker run -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=minio \
  -e MINIO_ROOT_PASSWORD=minio123 \
  minio/minio server /data --console-address ":9001"
````

---

### 2. PostgreSQL

Stores metadata, buckets, users, and analytics data.

```bash
docker run -p 5432:5432 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  postgres:16
```

---

### 3. NATS Server

Handles all system events.

```bash
docker run -p 4222:4222 nats:latest
```

---

## Environment Configuration

Create a `.env` file:

```env
APP_PROFILE=dev
SERVER_PORT=8082

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=s3

NATS_URL=nats://localhost:4222

S3_ENDPOINT=localhost:9000
S3_ACCESS_KEY=minio
S3_SECRET_KEY=minio123
S3_USE_SSL=false
```

---

## Running the Service

Start the Go service:

```bash
go run main.go
```

Or build and run:

```bash
go build -o s3-service
./s3-service
```

---

## Verification

Check service health:

```bash
curl http://localhost:8082/health/ping
```

Expected response:

```json
{
  "status": "ok"
}
```

---

## Common Issues

### 1. Port conflicts

Ensure ports are free:

* 8082 (API)
* 5432 (Postgres)
* 9000 (MinIO)
* 4222 (NATS)

---

### 2. MinIO connection failure

Check:

* credentials match `.env`
* container is running
* endpoint is correct

---

### 3. NATS not receiving events

Verify:

```bash
nats-server logs
```

---

## Development Notes

* Hot reload is not enabled by default
* Logs are structured via `slog`
* Event flow is asynchronous via NATS
* All storage operations depend on MinIO availability

---