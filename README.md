# Court

Automatically turn Kubernetes failures into actionable issues and AI-assisted fixes.

Court reduces the time between **failure detection** and **fix delivery** by catching runtime signals, structuring them as incidents, and routing them directly to your codebase.

When something fails in your cluster, Court:

- captures context from `kubectl describe` and logs
- build a structured `IncidentReport`
- routes it to the correct repository
- creates a ready-to-use issue in your VCS
- enables possibility to AI-assisted fixes (e.g. via GitHub Copilot)

From failure to fix in a single flow.

---

## Getting Started

TBD

---

## How it works

It is built around the following components:

- **Officer** -> detects problems inside the cluster
- **Docket** -> redis event bus
- **Court** -> processes incidents and drives resolution
- **Archive** -> postgreSQL persistence layer

---

## Motivation

Court turns runtime signals into:

- structured **incidents**
- tracked **lifecycles (Suits)**
- actionable **VCS issues**

Instead of manually debugging, copying logs, and opening issues, Court does it automatically.

---

## Core Concepts

### `IncidentReport`

Represents a failure detected in the cluster.

Built from:

- Pod metadata
- K8s events
- container logs, events, and signals

### `Suit`

Tracks the lifecycle of an incident.

States:

- `open`
- `closed`

A Suit:

- references an Incident
- links to VCS issue
- reflects real system state

---

## Lifecycle

1. A Pod fails
2. Officer detects the failure
3. Context is collected
4. Incident Report is created and persisted
5. Event is sent through Docket (redis)
6. Court processes it
7. A VCS issue is created
8. Developers (or AI agents) act on it
9. When resolved -> Suit and issue are closed

---

## Repository Resolution

Court determines where to open an issue:

1. **Pod annotation**

    ```yaml
    court.dev/repository: https://github.com/org/repo
    ```

2. **OCI image label**

    ```yaml
    org.opencontainers.image.source=https://github.com/org/repo
    ```

If neither is present, the incident is ignored.

---

## Contributing

Feedback is welcome via Issues and Discussions.

This project is in an early stage ideas, experiments, and contributions can have a real impact.

---

## License

Licensed under the Apache License 2.0.

---

Thank you ! 👩‍⚖️👨‍⚖️
