---
title: GameLift Overview
description: Introduction to Serwin GameLift, its architecture, core components, and how it powers scalable multiplayer game server infrastructure.
icon: cloud
tags: [gamelift, overview, architecture, multiplayer, infrastructure]
---

# GameLift Overview

Serwin GameLift is a managed service for deploying, operating, and scaling dedicated game servers in the cloud. It handles the infrastructure complexity so you can focus on your game.

---

## What GameLift Does

GameLift runs your game server builds on managed fleets of virtual machines. When a player wants to start or join a game, GameLift automatically finds an available server, creates a game session, and hands the player a connection endpoint — all in milliseconds.

---

## Key Capabilities

- **Dedicated servers**: Full server-side authority, no peer-to-peer. Your compiled game server binary runs on GameLift-managed VMs.
- **Auto scaling**: Fleets scale up and down based on player demand and custom scaling policies.
- **Matchmaking**: Built-in FlexMatch rules engine matches players by skill, latency, or any custom attribute.
- **Multi-region**: Deploy fleets across multiple regions and route players to the lowest-latency location.
- **Session management**: GameLift tracks every active game session, its player slots, and connection details.

---

## How It Fits Together

```

Player Client
│
▼
GameLift API  ──►  Matchmaking (FlexMatch)
│
▼
Fleet (VM pool)
│
▼
Game Server Process  ──►  Game Session  ──►  Player Sessions

```id="gamelift-architecture"

Your game server binary integrates the **GameLift Server SDK**, which handles process registration, health reporting, and session lifecycle callbacks. Your client integrates the **GameLift Client SDK** (or calls the REST API) to request matchmaking or create sessions directly.

---

## Core Components

| Component | Description |
|---|---|
| **Fleet** | A pool of VMs running your game server build |
| **Build** | Your compiled game server binary uploaded to GameLift |
| **Game Session** | A single running instance of your game (one match, one lobby) |
| **Player Session** | A reserved slot for one player inside a game session |
| **Alias** | A named pointer to a fleet — swap fleets with zero client changes |
| **Queue** | Routes session placement requests across multiple fleets/regions |
| **FlexMatch** | Matchmaking engine with configurable rule sets |

---

## Pricing Model

You pay for the underlying VM hours in your fleet.

- **On-Demand fleets**: Guaranteed capacity and stability for production workloads
- **Spot fleets**: Up to 70% cheaper, but can be interrupted with short notice

A common production setup uses a **hybrid model** with both for cost efficiency and reliability.
```
 