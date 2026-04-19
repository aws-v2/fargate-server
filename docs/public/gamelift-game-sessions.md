---
title: Game Sessions
description: Understand how game sessions are created, managed, and connected in GameLift, including lifecycle, matchmaking, and player sessions.
icon: gamepad
tags: [gamelift, sessions, matchmaking, multiplayer, realtime]
---

# Game Sessions

A game session is a single running instance of your game — one match, one lobby, one world. GameLift manages the full lifecycle from creation through termination and exposes connection details so players can join.

---

## Session Lifecycle

```

ACTIVATING → ACTIVE → TERMINATING → TERMINATED
│
ERROR

````id="sess-lifecycle"

- **ACTIVATING**: GameLift has assigned the session to a server process; the process is initializing.
- **ACTIVE**: The server has called `ActivateGameSession()` via the SDK. Players can now join.
- **TERMINATING**: The session is winding down (all players left, or the server called `TerminateGameSession()`).
- **TERMINATED**: The process slot is freed and available for the next session.

---

## Creating a Game Session

### Direct placement (single fleet)

```bash id="direct-session"
serwin gamelift create-game-session \
  --fleet-id fleet-xyz \
  --maximum-player-session-count 16 \
  --name "ranked-match-7391" \
  --game-properties '[
    {"Key":"map","Value":"dust2"},
    {"Key":"mode","Value":"competitive"}
  ]'
````

Game properties are forwarded to your server process in the `OnStartGameSession` callback and control runtime behavior such as map, mode, or tick rate.

---

### Queue-based placement (recommended)

Queue-based placement distributes sessions across multiple fleets and regions for better latency and availability.

```bash id="queue-placement"
serwin gamelift start-game-session-placement \
  --placement-id "placement-$(uuidgen)" \
  --game-session-queue-name prod-queue \
  --maximum-player-session-count 16 \
  --player-latencies '[
    {"PlayerId":"player-1","RegionIdentifier":"us-east-1","LatencyInMilliseconds":12},
    {"PlayerId":"player-1","RegionIdentifier":"eu-west-1","LatencyInMilliseconds":95}
  ]'
```

Check status:

```bash id="placement-status"
serwin gamelift describe-game-session-placement \
  --placement-id "placement-abc123"
```

When status becomes `FULFILLED`, the response contains the connection endpoint.

---

## Player Sessions

A player session reserves a slot in a game session for a specific user.

```bash id="player-session"
serwin gamelift create-player-session \
  --game-session-id gsess-abc \
  --player-id "user-42"
```

The returned `PlayerSessionId` is sent to the client and validated on connection.

Server-side validation:

```go
gamelift.AcceptPlayerSession(playerSessionId)
```

On disconnect:

```go
gamelift.RemovePlayerSession(playerSessionId)
```

This ensures slot consistency and prevents unauthorized connections.

---

## FlexMatch Matchmaking

FlexMatch is GameLift’s matchmaking engine for creating balanced multiplayer sessions using rule-based logic.

---

### Rule Set Example

```json id="flexmatch-rule-set"
{
  "name": "two-team-competitive",
  "ruleLanguageVersion": "1.0",
  "playerAttributes": [
    { "name": "skill", "type": "number" }
  ],
  "teams": [
    { "name": "red",  "maxPlayers": 5, "minPlayers": 5 },
    { "name": "blue", "maxPlayers": 5, "minPlayers": 5 }
  ],
  "rules": [
    {
      "name": "skill-balance",
      "type": "distance",
      "measurements": ["teams[red].players.attributes[skill]"],
      "referenceValue": "teams[blue].players.attributes[skill]",
      "maxDistance": 300
    },
    {
      "name": "latency-cap",
      "type": "latency",
      "maxLatency": 150
    }
  ],
  "expansions": [
    {
      "target": "rules[skill-balance].maxDistance",
      "steps": [
        { "waitTimeSeconds": 10, "value": 500 },
        { "waitTimeSeconds": 20, "value": 1000 }
      ]
    }
  ]
}
```

Expansions prevent long queue times by relaxing constraints over time.

---

### Start Matchmaking

```bash id="start-matchmaking"
serwin gamelift start-matchmaking \
  --configuration-name two-team-competitive \
  --players '[
    {
      "PlayerId": "player-1",
      "PlayerAttributes": { "skill": {"N": 1450} },
      "LatencyInMs": { "us-east-1": 20, "eu-west-1": 110 }
    }
  ]'
```

Poll:

```bash id="poll-matchmaking"
serwin gamelift describe-matchmaking \
  --ticket-id matchmaking-ticket-123
```

Status `COMPLETED` returns session and team assignment data.

---

## Session Search & Browsing

Used for open lobbies where players manually join games.

```bash id="search-sessions"
serwin gamelift search-game-sessions \
  --fleet-id fleet-xyz \
  --filter-expression "hasAvailablePlayerSessions=true" \
  --sort-expression "creationTimeMillis ASC" \
  --limit 20
```

Supports filtering by built-in and custom properties.

---

## Connection Flow

````
1. Client requests matchmaking/session placement
2. GameLift assigns session + returns IP/Port
3. Client connects to server
4. Client sends PlayerSessionId
5. Server validates via AcceptPlayerSession
6. Gameplay begins
``` id="connection-flow"

---

## Key Rules

- Never expose fleet IDs to clients
- Always validate player sessions server-side
- Use queue-based placement for production
- Treat sessions as ephemeral compute units
```` 