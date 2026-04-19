---
title: "Storage Engine Design"
description: "How the S3 service uses MinIO as its underlying object storage engine."
icon: "storage"
tags: ["storage", "minio", "architecture"]
---

# 🧠 Storage Engine Design

The S3 service uses **MinIO as the primary object storage engine**.

All file data is stored in MinIO, while metadata is stored in PostgreSQL.

---

# 🏗 Storage Architecture

```

Client Upload
↓
HTTP Handler
↓
Application Service
↓
Storage Adapter Layer
↓
MinIO Client
↓
Object Stored in Bucket

```id="se1"

---

# 🔌 Storage Abstraction Layer

The system does NOT interact with MinIO directly from handlers.

Instead it uses:

```

MinIOAdapter

```id="se2"

Located in:

```

/internal/infrastructure/storage

```id="se3"

### Responsibilities:
- upload objects
- download objects
- delete objects
- generate presigned URLs
- manage bucket operations

---

# 🧾 Why MinIO is Abstracted

This design allows:

- swap to AWS S3 without code changes
- testable storage layer
- separation of concerns
- clean service boundaries

---

# 📦 Data Split Model

| Layer        | Responsibility |
|--------------|---------------|
| MinIO        | Raw file bytes |
| PostgreSQL   | Metadata       |
| NATS         | Events         |

---

# 🔄 Write Flow

1. Client uploads file
2. Service validates request
3. File sent to MinIO
4. Metadata saved in DB
5. Event emitted to NATS

---

# 📌 Key Rule

> MinIO stores data. It does NOT store business logic or metadata.

---

# ⚠ Failure Handling

If MinIO upload fails:
- DB write is aborted
- event is NOT emitted
- request returns error immediately
```

---

