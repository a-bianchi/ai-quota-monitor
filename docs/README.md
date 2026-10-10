# Documentation

**English** · [Español](README.es.md)

Index of the AI Quota Monitor documentation. For an overview, start with the [main README](../README.md).

## Getting started

| Topic | Document |
|---|---|
| Install and run on your machine | [Local setup](local-setup.md) |
| Docker Compose, homelab, Coolify | [Deployment](deployment.md) |
| Environment variables, tokens, and personal and work accounts | [Configuration](configuration.md) |
| Common problems | [Troubleshooting](troubleshooting.md) |

## Using it

| Topic | Document |
|---|---|
| Web panel: user, session and credentials | [Web panel](panel.md) |
| Connecting each provider step by step | [Connection guide](providers/connect.md) |
| Status and scope of each integration | [Providers](providers/README.md) |
| Codex (ChatGPT) | [Codex](providers/codex.md) |
| Cursor | [Cursor](providers/cursor.md) |
| Claude Pro/Max | [Claude](providers/claude-subscription.md) |

## Reference

| Topic | Document |
|---|---|
| Endpoints, quota contract, units, statuses and `usageLevel` | [API](api.md) |
| Postman collection and environment | [`collections/`](collections/) |
| Components, data on disk and code layout | [Architecture](architecture.md) |

## Hardware

| Topic | Document | Status |
|---|---|---|
| Board, materials and pins | [Hardware](hardware.md) | Official demo tested on the board |
| Firmware: step-by-step setup, flashing and troubleshooting | [Firmware](firmware.md) | Compiles; not tested on the board |
| Enclosure and 3D printing | [Enclosure](../enclosure/README.md) | Pending |

## Project

| Topic | Document |
|---|---|
| What is done and what is missing | [Roadmap](roadmap.md) |
| How to contribute and add a provider | [CONTRIBUTING](../CONTRIBUTING.md) |
| License | [MIT](../LICENSE) |

The screenshots in [`assets/`](assets/) were taken from the real panel with demo data. [Web panel](panel.md#screenshots) explains how to regenerate them.
