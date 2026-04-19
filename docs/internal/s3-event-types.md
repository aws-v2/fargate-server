---
title: Event Types
description: Complete catalog of all NATS events emitted by the S3 service across object, bucket, security, and system domains
icon: list-bullet
tags: [events, nats, domain-events, messaging, system-design]
---

# 📘 `s3-event-types.md`

## Event Types

### Overview

All system events are published via NATS subjects using a consistent naming convention:

```

<domain>.<action>

```

Example:

```

object.uploaded
bucket.created
security.access_denied

````

This naming structure ensures:

- Clear domain ownership
- Predictable event routing
- Easy consumer subscription patterns

---

## 1. Object Events

### `object.uploaded`

Triggered after a successful upload to MinIO.

**Payload:**

```json
{
  "bucket_id": "bucket-1",
  "object_key": "file.png",
  "size": 12345,
  "content_type": "image/png"
}
````

---

### `object.deleted`

Triggered when an object is permanently removed from storage.

---

### `object.moved`

Triggered when an object is moved or renamed between:

* prefixes
* folders
* buckets

---

## 2. Bucket Events

### `bucket.created`

Emitted immediately after a new bucket is successfully created.

---

### `bucket.deleted`

Emitted when a bucket is removed from the system.

---

### `bucket.updated`

Emitted when any bucket configuration changes, including:

* lifecycle rules
* encryption settings
* versioning state
* access policies

---

## 3. Access Events

### `access.presigned.created`

Triggered when a presigned URL is generated for:

* upload
* download
* multipart operations

---

### `access.presigned.used`

Triggered when a presigned URL is successfully used to access an object.

---

## 4. Security Events

### `security.access_denied`

Triggered when a request is rejected by IAM due to:

* missing permissions
* invalid API key
* invalid JWT claims
* policy restrictions

---

### `security.policy_updated`

Triggered when bucket or user policies are modified.

---

## 5. System Events

### `system.health_check`

Emitted for internal health monitoring and observability.

---

### `system.startup`

Emitted once when the service successfully starts and completes initialization.

---

## Design Rules

These rules apply to all events in the system:

* Events must remain **backward compatible**
* Never delete or rename existing event types (only deprecate)
* Payload structure must remain **schema-stable**
* Consumers must safely ignore unknown fields
* Events are **at-least-once delivery**, so consumers must be idempotent

---

## Architectural Note

Events are a **core system contract**, not just logs.

They power:

* Analytics pipelines
* Metrics aggregation
* Security monitoring
* Future distributed features (replication, indexing, AI analysis)

---

 
