---
title: Fleets
description: Understand how fleets power your game server infrastructure, including lifecycle, scaling, instance types, and deployment strategies.
icon: server
tags: [gamelift, fleets, infrastructure, scaling, compute]
---

# Fleets

A fleet is a pool of virtual machine instances that run your game server build. Fleets are the core compute unit in GameLift — everything else (game sessions, matchmaking, queues) ultimately places workloads onto a fleet.

## Fleet Lifecycle

```

NEW → DOWNLOADING → VALIDATING → BUILDING → ACTIVATING → ACTIVE
│
ERROR ◄─────────┤
DELETING ◄─────────┘

````

A fleet only accepts game sessions when it is in the `ACTIVE` state. If activation fails, check your runtime configuration and server process launch path.

## Instance Types

Choose an instance type based on your game server's CPU and memory requirements:

| Type | vCPUs | Memory | Typical Use Case |
|---|---|---|---|
| `c5.large` | 2 | 4 GB | Small sessions, turn-based games |
| `c5.xlarge` | 4 | 8 GB | FPS / action games, ~50 players |
| `c5.2xlarge` | 8 | 16 GB | Large lobbies, simulation games |
| `m5.xlarge` | 4 | 16 GB | Memory-heavy game worlds |

Benchmark your server process under peak load to pick the right type. Over-provisioning is expensive; under-provisioning causes latency spikes.

## On-Demand vs Spot Fleets

**On-Demand** fleets guarantee instance availability. Use them for production workloads where interruption is unacceptable.

**Spot** fleets use spare capacity at up to 70% lower cost, but instances can be reclaimed with a two-minute warning. GameLift handles the warning by draining active sessions from the instance before termination. Use spot fleets for:
- Development and testing environments
- Non-critical overflow capacity behind a queue

A common production pattern is a **mixed queue**: a spot fleet as the primary destination and an on-demand fleet as the fallback.

## Runtime Configuration

The runtime config defines which server processes run on each VM and how many run concurrently:

```json
{
  "ServerProcesses": [
    {
      "LaunchPath": "/local/game/MyGameServer",
      "Parameters": "-port 7777 -mode competitive",
      "ConcurrentExecutions": 1
    },
    {
      "LaunchPath": "/local/game/MyGameServer",
      "Parameters": "-port 7778 -mode casual",
      "ConcurrentExecutions": 3
    }
  ],
  "MaxConcurrentGameSessionActivations": 2,
  "GameSessionActivationTimeoutSeconds": 300
}
````

`ConcurrentExecutions` across all process entries must not exceed the VM's capacity. `MaxConcurrentGameSessionActivations` throttles how many sessions can activate in parallel to prevent CPU spikes during burst.

## Auto Scaling

GameLift provides two scaling mechanisms:

### Target-Based Scaling (recommended)

Maintain a target percentage of available game sessions:

```bash
serwin gamelift put-scaling-policy \
  --fleet-id fleet-xyz \
  --name "maintain-20pct-buffer" \
  --policy-type TargetBased \
  --target-configuration '{"TargetValue": 20}' \
  --metric-name PercentAvailableGameSessions
```

This keeps 20% of your sessions free, so new players are never waiting for a VM to spin up.

### Rule-Based Scaling

Trigger scale-out/in actions based on CloudWatch metrics or custom thresholds:

```bash
serwin gamelift put-scaling-policy \
  --fleet-id fleet-xyz \
  --name "scale-out-on-full" \
  --scaling-adjustment 2 \
  --scaling-adjustment-type ChangeInCapacity \
  --threshold 90 \
  --comparison-operator GreaterThanThreshold \
  --metric-name PercentAvailableGameSessions \
  --evaluation-periods 1
```

### Scaling Limits

Always set minimum and maximum capacity to bound your costs:

```bash
serwin gamelift update-fleet-capacity \
  --fleet-id fleet-xyz \
  --min-size 2 \
  --max-size 50 \
  --desired-instances 5
```

## Fleet Aliases

An alias is a named pointer to a fleet. Clients and queues reference the alias, not the fleet ID directly. This lets you swap out a fleet (e.g. after uploading a new build) with no client-side changes:

```bash
# Create alias
serwin gamelift create-alias \
  --name "prod-game-server" \
  --routing-strategy '{"Type": "SIMPLE", "FleetId": "fleet-xyz"}'

# Later — point alias to new fleet, zero downtime
serwin gamelift update-alias \
  --alias-id alias-abc \
  --routing-strategy '{"Type": "SIMPLE", "FleetId": "fleet-new"}'
```

## Fleet Locations

By default a fleet runs in your home region. Add remote locations to reduce latency for players in other regions:

```bash
serwin gamelift create-fleet-locations \
  --fleet-id fleet-xyz \
  --locations '[{"Location":"eu-west-1"},{"Location":"ap-southeast-1"}]'
```

GameLift replicates your build to those locations automatically and manages capacity independently per location.

``` 