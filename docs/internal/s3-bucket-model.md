---
title: "Bucket Model"
description: "Defines how buckets are structured, owned, and managed in the S3 system."
icon: "bucket"
tags: ["bucket", "multi-tenancy", "storage"]
---

# 🪣 Bucket Model

Buckets are the **top-level container for all objects**.

Each bucket belongs to a single user (multi-tenant model).

---

# 👤 Ownership Model

```

User → Bucket → Objects

````id="bm1"

Each bucket contains:

- ownerId
- metadata
- policies
- configuration settings

---

# 🧠 Bucket Structure

Stored in PostgreSQL:

```json id="bm2"
{
  "bucketId": "uuid",
  "name": "my-bucket",
  "ownerId": "user-123",
  "createdAt": "timestamp",
  "versioning": true,
  "encryption": true
}
````

---

# 🔐 Access Control

All bucket operations are protected by:

````
IAM Middleware → Actor Context → Bucket Ownership Check
``` id="bm3"

Rules:

- only owner can modify bucket
- IAM may grant elevated access
- internal services may bypass via middleware

---

# ⚙ Bucket Lifecycle

Buckets support:

- creation
- deletion
- emptying
- policy updates
- versioning toggle
- lifecycle rules
- encryption config

---

# 🔄 Bucket Isolation Model

Each bucket is logically isolated:

- no cross-bucket file access
- no shared namespace
- no global listing across users

---

# 📌 Design Principle

> Buckets are tenant boundaries in the system.

They define security + data isolation.
````

---
