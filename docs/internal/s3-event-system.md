---
title: Event-Driven Architecture (NATS)
description: Core event system powering decoupled storage operations, analytics, and system observability across the S3 service
icon: lightning-bolt
tags: [events, nats, architecture, async, system-design]
---

## Event-Driven Architecture (NATS)

### Overview

The S3 service uses **NATS as the central event bus** to decouple core storage operations from side effects such as:

* Metrics collection
* Audit logging
* Analytics processing
* IAM-related enforcement hooks
* Future extensibility (replication, notifications, etc.)

Instead of tightly coupling logic inside HTTP handlers or services, the system emits **domain events** that are consumed asynchronously.

---

### Why NATS?

We use NATS because it provides:

* Low-latency pub/sub messaging
* Lightweight operational overhead
* High throughput for storage events
* Simple mental model for microservices
* Native support for fan-out event consumption

In this system:

> **Every important mutation in S3 emits an event.**

---

### Event Flow Architecture
