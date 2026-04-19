---
title: Quickstart: Deploy a Game Server
description: Step-by-step guide to uploading a game server, creating a fleet, and launching your first GameLift game session.
icon: rocket
tags: [gamelift, quickstart, deployment, fleet, server-setup]
---

# Quickstart: Deploy a Game Server

This guide walks you through uploading a game server build, creating a fleet, and starting your first game session.

---

## Prerequisites

- A compiled game server binary integrated with the GameLift Server SDK
- Serwin CLI installed and authenticated
- A GameLift-enabled account

---

## Step 1 — Upload Your Build

Package your game server binary and dependencies, then upload it to GameLift:

```bash id="upload-build"
serwin gamelift upload-build \
  --name "MyGame-Server-v1.0" \
  --version "1.0.0" \
  --build-root ./server-build/ \
  --operating-system AMAZON_LINUX_2
````

You will receive a `BuildId` (e.g. `build-abc123`). Store it for the next step.

---

## Step 2 — Create a Fleet

A fleet is a pool of virtual machines that run your game server build:

```bash id="create-fleet"
serwin gamelift create-fleet \
  --name "MyGame-Fleet-Dev" \
  --build-id build-abc123 \
  --ec2-instance-type c5.large \
  --runtime-config '{
    "ServerProcesses": [
      {
        "LaunchPath": "/local/game/MyGameServer",
        "Parameters": "-port 7777 -logPath /local/game/logs",
        "ConcurrentExecutions": 2
      }
    ]
  }'
```

`ConcurrentExecutions` defines how many server processes run per VM.

Fleet lifecycle:

````
NEW → DOWNLOADING → ACTIVATING → ACTIVE
``` id="fleet-lifecycle"

Wait until status becomes `ACTIVE` before creating sessions:

```bash id="fleet-status"
serwin gamelift describe-fleet-attributes --fleet-id fleet-xyz789
````

---

## Step 3 — Start a Game Session

Once the fleet is active, create your first game session:

```bash id="create-session"
serwin gamelift create-game-session \
  --fleet-id fleet-xyz789 \
  --maximum-player-session-count 10 \
  --name "TestMatch-001"
```

Response includes:

* `IpAddress`
* `Port`

Clients use these to connect.

---

## Step 4 — Reserve a Player Slot

Reserve a slot before a player connects:

```bash id="player-slot"
serwin gamelift create-player-session \
  --game-session-id gsess-abc \
  --player-id "player-42"
```

Pass the returned `PlayerSessionId` to the client.

The server validates it during connection.

---

## Step 5 — Integrate the Server SDK

Minimal Go integration example:

```go id="sdk-example"
import "github.com/serwin/gamelift-server-sdk-go"

func main() {
    gamelift.InitSDK()
    defer gamelift.Destroy()

    processParams := gamelift.ProcessParameters{
        OnStartGameSession: func(session gamelift.GameSession) {
            gamelift.ActivateGameSession()
        },
        OnProcessTerminate: func() {
            gamelift.ProcessEnding()
        },
        HealthCheck: func() bool {
            return true
        },
        Port: 7777,
        LogParameters: gamelift.LogParameters{
            LogPaths: []string{"/local/game/logs/server.log"},
        },
    }

    gamelift.ProcessReady(processParams)
}
```

---

## What’s Next

* Set up **auto scaling** → see [Fleets](gamelift-fleets)
* Configure **matchmaking rules** → see [Game Sessions](gamelift-game-sessions)
* Use **queues for multi-region placement**

```

--- 