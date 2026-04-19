---
title: "Startup Lifecycle"
description: "How the S3 service boots and initializes all dependencies."
icon: "startup"
tags: ["startup", "boot", "lifecycle"]
---

# 🚀 Startup Lifecycle

This document explains how the system starts.

---

# 🧭 Boot Sequence

## 1. Load Configuration

```

cfg, err := config.Load()

```id="st1"

Loads:

- database config
- NATS config
- S3 config
- Eureka config

---

## 2. Logger Initialization

```

logging.InitLogger(cfg.APP_PROFILE)

```id="st2"

Sets environment-aware logging:
- dev → debug logs
- prod → structured logs

---

## 3. Service Registration (Eureka)

```

registerWithEureka()

```id="st3"

Registers service instance for discovery.

---

## 4. Start Heartbeat

```

go sendHeartbeat()

```id="st4"

Keeps service alive in Eureka registry.

---

## 5. Reachability Checks

System verifies:

- NATS reachable
- PostgreSQL reachable
- MinIO reachable

If any fail → service exits.

---

## 6. Initialize Storage

```

minioAdapter := storage.NewMinIOAdapter()

```id="st5"

Connects to object storage backend.

---

## 7. Initialize NATS

```

natsAdapter := event.NewNATSAdapter()

```id="st6"

Used for:
- events
- metrics
- IAM communication

---

## 8. Database Connection

```

database.NewPostgresDB()

```id="st7"

Used for:
- metadata
- bucket info
- file records

---

## 9. Run Migrations

```

database.RunMigrations()

```id="st8"

Ensures schema is up to date.

---

## 10. Initialize Services

All application services are created:

- BucketService
- UploadService
- PresignService
- AnalyticsService

---

## 11. Initialize HTTP Layer

```

router := gin.Default()
RegisterRoutes(router, handlers)

```id="st9"

---

## 12. Start Server

```

router.Run(":8082")

```id="st10"

Service is now live.

---

# 🔄 Startup Flow Diagram

```

Config
↓
Logger
↓
Eureka Registration
↓
Heartbeat
↓
Health Checks
↓
MinIO + NATS + DB Init
↓
Migrations
↓
Services Init
↓
HTTP Server Start

```id="st11"

---

# 📌 Key Insight

This startup order ensures:

- no partial startup states
- dependency safety
- early failure detection
```

--- 