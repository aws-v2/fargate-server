
---
title: "Core Components Breakdown"
description: "Breakdown of all major internal modules and their responsibilities."
icon: "components"
tags: ["architecture", "components", "backend"]
---

# 🧩 Core Components

The system is structured into **4 main layers**:

---

# 1. HTTP Layer (Transport)

Location:
```

/internal/transport/http

```id="cmp1"

### Responsibilities:
- request routing (Gin)
- input validation
- auth context extraction
- response formatting

### Example:
```

BucketHandler.CreateBucket()
FileHandler.UploadFile()

```id="cmp2"

---

# 2. Application Layer (Business Logic)

Location:
```

/internal/application

```id="cmp3"

### Responsibilities:
- orchestrates business logic
- enforces rules
- coordinates storage + DB + events

Example services:

- BucketService
- UploadService
- PresignService

---

# 3. Infrastructure Layer

Location:
```

/internal/infrastructure

```id="cmp4"

### Responsibilities:
- MinIO adapter
- PostgreSQL repository
- NATS client
- logging system
- metrics system

This layer contains all external dependencies.

---

# 4. Middleware Layer

Location:
```

/internal/middleware

```id="cmp5"

### Responsibilities:
- authentication (JWT / API key)
- IAM validation
- actor injection
- request tracking

---

# 🔄 Component Interaction Flow

```

HTTP Handler
↓
Application Service
↓
Infrastructure (DB / Storage / NATS)

```id="cmp6"

---

# 📌 Key Design Rule

> Handlers must NEVER call infrastructure directly.

Everything must go through the application layer.
```

---

