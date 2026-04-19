---
title: "Authorization Model"
description: "How access control is enforced for buckets, files, and system operations."
icon: "shield"
tags: ["authorization", "iam", "security", "rbac"]
---

# 🛡 Authorization Model

Authorization determines what an authenticated actor can do.

---

# 🧠 Core Model

The system uses a **hybrid model**:

- RBAC (Role-Based Access Control)
- Resource Ownership Model
- Service-level bypass rules

---

# 👤 Actor-Based Access

Every request has an actor:

```

actor.ID
actor.Roles

```id="authz1"

---

# 🪣 Bucket Authorization

Buckets are **owner-scoped resources**:

### Rule:
```

actor.ID must match bucket.ownerId

```

---

# 📦 Object Authorization

Objects inherit bucket permissions:

```

Object → Bucket → Owner

```id="authz2"

If you can access bucket → you can access objects inside.

---

# 🔐 IAM Validation Flow

```

Handler
↓
Service
↓
IAM Validator (NATS)
↓
Allow / Deny

```id="authz3"

---

# 🚪 Permission Types

System supports:

- READ
- WRITE
- DELETE
- ADMIN

---

# ⚙ Special Cases

## 1. Internal Services

Some services bypass IAM:

- analytics
- metrics
- system controllers

Controlled via:

```

LocalBypassMiddleware

```id="authz4"

---

## 2. Presigned URLs

Authorization is embedded in:

- signed token
- expiry time
- bucket scope

No IAM call needed at runtime.

---

# ⚠ Security Rule

> IAM must be enforced before any storage or DB operation.

Never trust handler-level checks alone.

---

# 📌 Design Principle

- Authentication → who you are
- Authorization → what you can do
- Actor → runtime identity context
```

---

# 🚀 You now have IAM fully documented

✔ architecture
✔ auth flow
✔ authorization model
✔ middleware behavior
✔ actor system
✔ NATS-based validation
✔ bypass rules

---

 
