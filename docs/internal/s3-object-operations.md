
# 📘 3. Object Operations Flow

```md id="obj1"
---
title: "Object Operations Flow"
description: "How file operations flow through the system from API to MinIO and back."
icon: "object"
tags: ["objects", "upload", "download", "flow"]
---

# 📦 Object Operations Flow

This document explains how file operations work end-to-end.

---

# 🔄 Supported Operations

- Upload file
- Download file
- Delete file
- Copy file
- Move file
- Update metadata

---

# 🧭 Upload Flow (Most Important)

```

Client
↓
HTTP Handler
↓
Auth Middleware (IAM)
↓
Upload Service
↓
MinIO Adapter
↓
PostgreSQL (metadata)
↓
NATS event: file.uploaded

```id="of1"

---

# 📤 Upload Steps

## 1. Request validation
- bucketId extracted
- file validated
- size/type checked

## 2. IAM check
- verify actor ownership
- validate permissions

## 3. Storage write
File stored in MinIO:

```

bucketName/objectKey

```id="of2"

## 4. Metadata persistence
Stored in PostgreSQL:

- fileId
- bucketId
- ownerId
- size
- timestamps

## 5. Event emission
NATS event published:

```

file.uploaded

```

---

# 📥 Download Flow

```

Client → Handler → Service → MinIO → Stream response

```id="of3"

Metadata is NOT modified.

---

# ❌ Delete Flow

```

Client → IAM → Service → MinIO delete → DB delete → event

```id="of4"

Ensures consistency between storage and metadata.

---

# 🔁 Copy / Move Flow

### Copy:
- duplicate object in MinIO
- create new metadata entry

### Move:
- copy object
- delete original
- update metadata references

---

# ⚠ Consistency Model

The system is **eventually consistent** for events, but:

- storage + DB operations are strongly consistent within request scope

---

# 📌 Key Rule

> MinIO is source of truth for file bytes  
> PostgreSQL is source of truth for file metadata
```

---

# 🚀 You now have Storage Layer documentation that is:

✔ production-grade
✔ aligned with your actual code
✔ explains real flows (not theory)
✔ onboarding-ready
✔ event + IAM aware

---
 