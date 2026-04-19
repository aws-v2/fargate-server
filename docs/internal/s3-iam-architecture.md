---
title: "IAM Architecture"
description: "How IAM is structured across middleware, NATS, and service boundaries."
icon: "lock"
tags: ["iam", "security", "architecture", "auth"]
---

# 🔐 IAM Architecture Overview

The S3 service uses a **distributed IAM model** enforced at the middleware layer and validated via NATS.

---

# 🧠 Core Idea

IAM is NOT a standalone service inside this repo.

Instead it is composed of:

- HTTP Middleware (request interception)
- IAM Validator (NATS-backed)
- Actor Context propagation
- Service-level enforcement

---

# 🏗 IAM System Design

```

Client Request
↓
BearerAuthMiddleware
↓
Extract Actor (user/service)
↓
IAM Validator (NATS call)
↓
Attach Actor to Context
↓
Handler / Service Layer

```id="iam_arch_1"

---

# 🔌 Components

## 1. Middleware Layer

Located in:
```

/internal/middleware

````id="iam_arch_2"

Responsibilities:
- extract JWT / API key
- attach actor to request context

---

## 2. IAM Validator

```go
middleware.NewIAMValidator(natsConnection)
``` id="iam_arch_3"

Responsibilities:
- validate permissions
- check access policies
- respond via NATS request/reply

---

## 3. Actor Context

Every request carries:

```go id="iam_arch_4"
type Actor struct {
    ID string
    Roles []string
}
````

Injected into Gin context:

```
c.Set("actor", actor)
```

---

# 🔄 IAM Flow Principle

> IAM is enforced at the edge (middleware), not inside services.

Services assume:

* actor is already validated
* permissions are already checked (unless explicitly required)

---

# 📌 Design Goal

* centralized authentication
* distributed authorization
* stateless services
* event-driven policy evaluation

````

---
 