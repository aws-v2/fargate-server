---
title: "Architecture Overview"
description: "High-level system architecture of the S3 microservice including storage, IAM, and event-driven design."
icon: "architecture"
tags: ["architecture", "system", "design"]
---

# 🧠 System Architecture Overview

The S3 service is a **multi-tenant object storage microservice** designed around:

- Stateless HTTP API (Gin)
- MinIO object storage backend
- PostgreSQL metadata store
- NATS event-driven system
- IAM-based authorization layer

---

# 🏗 High-Level Architecture

```

Client
│
▼
API Gateway
│
▼
Gin HTTP Layer (S3 Service)
│
├── IAM Middleware (AuthZ)
├── Application Layer (Business Logic)
├── Metrics + Analytics (NATS)
│
▼
Storage Adapter Layer
│
▼
MinIO (Object Storage)

Parallel:
Gin Layer → NATS → Event Consumers (Analytics, Metrics, Auditing)

```id="arch1"

---

# 🔑 Core Principles

## 1. Stateless API Design
The HTTP layer contains no business state. All state is externalized:

- PostgreSQL → metadata
- MinIO → objects
- NATS → events

---

## 2. Event-Driven Architecture
Every mutation emits an event:

- file uploaded
- bucket created
- file deleted
- presigned URL generated

This allows:
- analytics
- auditing
- async processing

---

## 3. Storage Abstraction
MinIO is abstracted behind:

```

storage adapter interface

```id="arch2"

This allows future migration to AWS S3 without code changes.

---

## 4. Multi-Tenant Isolation
Every resource is scoped:

```

User → Buckets → Objects

```id="arch3"

No cross-user access unless explicitly authorized via IAM.

---

# 🔐 Security Model

All requests pass through:

```

BearerAuthMiddleware → IAM Validator → Actor Context

```id="arch4"

Each request is tied to an `actor`.

---

# 📌 Summary

This system is designed for:

- horizontal scaling
- event-driven extensibility
- strict tenant isolation
- storage backend flexibility
```

---
