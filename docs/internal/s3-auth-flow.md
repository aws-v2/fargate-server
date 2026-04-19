---
title: "Authentication Flow"
description: "How requests are authenticated using JWT/API keys and converted into actors."
icon: "key"
tags: ["authentication", "jwt", "middleware"]
---

# 🔑 Authentication Flow

Authentication happens at the **middleware layer** before any business logic executes.

---

# 🧭 Flow Overview

````

Request
↓
BearerAuthMiddleware
↓
Validate Token / API Key
↓
Resolve Actor Identity
↓
Attach Actor to Context
↓
Continue Request

```id="auth1"

---

# 🔐 Supported Auth Methods

## 1. JWT Authentication

Used for:
- user requests
- frontend applications

Flow:
- token parsed
- signature validated
- user ID extracted

---

## 2. API Key Authentication

Used for:
- internal services
- system-to-system calls

---

# 🧠 Actor Creation

After authentication:

```

actor := domain.Actor{
ID: userId,
}

```id="auth2"

Then stored in Gin context:

```

c.Set("actor", actor)

```

---

# ⚙ Middleware Chain

```

Analytics Middleware
↓
Auth Middleware
↓
IAM Middleware (optional validation)
↓
Handler

```id="auth3"

---

# 🚨 Failure Cases

### Invalid token
→ 401 Unauthorized

### Missing token
→ 401 Unauthorized

### Invalid actor format
→ 500 Internal Error (should not happen in production)

---

# 📌 Key Rule

> Authentication only identifies the user. Authorization decides access.
```

---