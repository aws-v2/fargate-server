---
title: "Request Lifecycle"
description: "Step-by-step explanation of how a request flows through the system."
icon: "flow"
tags: ["request", "lifecycle", "architecture"]
---

# 🔄 Request Lifecycle

This document explains how a request flows through the S3 system.

---

# 📥 Example Request

```

POST /api/v1/s3/files/upload/:bucketId

```id="req1"

---

# 🧭 Step-by-Step Flow

## 1. API Gateway
- routes request to S3 service
- adds forwarding headers

---

## 2. Gin Router

Request enters:

```

RegisterRoutes()

```id="req2"

Then:

```

registerObjectRoutes()

````id="req3"

---

## 3. Middleware Chain

Executed in order:

- Analytics middleware
- Auth middleware (if enabled)
- IAM validator

Actor is injected:

```go
c.Set("actor", Actor{})
``` id="req4"

---

## 4. Handler Layer

Example:

````

UploadFile()

```id="req5"

Responsibilities:
- validate request
- extract bucketId
- parse payload
- call service layer

---

## 5. Application Layer

```

UploadService.Upload()

```id="req6"

Responsibilities:
- validate business rules
- check bucket ownership
- coordinate storage + DB
- emit NATS event

---

## 6. Storage Layer (MinIO)

```

minioAdapter.PutObject()

```id="req7"

File is physically stored in MinIO.

---

## 7. Database Layer

Metadata stored:

- fileId
- bucketId
- ownerId
- size
- timestamps

---

## 8. Event Emission (NATS)

Event emitted:

```

file.uploaded

```id="req8"

Consumers:
- analytics service
- metrics service
- auditing (future)

---

# 📊 Full Flow Diagram

```

Client
↓
Gateway
↓
Gin Router
↓
Middleware (Auth + IAM)
↓
Handler
↓
Service Layer
↓
┌───────────────┬───────────────┐
▼               ▼               ▼
MinIO        PostgreSQL      NATS Events

```id="req9"

---

# ⚠ Important Rules

- handlers must stay thin
- services must contain all logic
- infrastructure must never call services
```

---

