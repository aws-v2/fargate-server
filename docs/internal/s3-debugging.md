---
title: Debugging Guide
description: Internal debugging guide for tracing requests, diagnosing failures, and inspecting event flows in the S3 service
icon: bug
tags: [debugging, troubleshooting, logs, observability, development]
---

# 📘 `s3-debugging.md`

## Debugging Guide

### Overview

This guide helps engineers debug issues across:

- API layer (Gin handlers)
- Application services
- MinIO storage operations
- NATS event flow
- Database interactions

---

## Logging System

The service uses structured logging via `slog`.

Example log:

```json id="log1"
{
  "time": "2026-04-18T22:00:00Z",
  "level": "INFO",
  "msg": "Object uploaded",
  "bucket_id": "bucket-1"
}
````

---

## Request Tracing

Every request flows through:

```
Handler → Service → Storage → Event → Response
```

To debug:

1. Check HTTP logs
2. Trace service layer logs
3. Verify MinIO operation
4. Confirm NATS event emission

---

## Common Debug Scenarios

### 1. Upload failing

Check:

* MinIO connectivity
* Bucket existence
* IAM permissions
* Payload size limits

---

### 2. Events not firing

Check:

* NATS connection
* Publisher in service layer
* Consumer subscriptions

---

### 3. Missing metadata

Check:

* PostgreSQL writes
* Transaction rollback
* Service layer mapping

---

## Debugging Tools

### 1. API Testing

```bash
curl -v http://localhost:8082/api/v1/s3/health/ping
```

---

### 2. NATS Monitoring

```bash
nats sub "object.*"
```

---

### 3. Database Inspection

```bash
psql -U postgres -d s3
```

---

### 4. MinIO Console

Access:

```
http://localhost:9001
```

---

## Debug Strategy

Follow this order:

1. HTTP layer (request received?)
2. Service layer (business logic executed?)
3. Storage layer (MinIO success?)
4. Event layer (NATS published?)
5. Consumer layer (metrics/analytics processed?)

---

## Key Principle

> Never assume failure is in one layer — trace end-to-end.

---
 