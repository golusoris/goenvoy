# AGENTS — downloadclient/queue

> Per-module notes. Read [../AGENTS.md](../AGENTS.md) first.

## Purpose

Unified read-only queue snapshot aggregator across all 6 download client modules (qBittorrent, Transmission, Deluge, rTorrent, SABnzbd, NZBGet).

## Operation

`Take(ctx, Target, opts...)` inspects client queue state. Read-only: never issues mutating commands (no pause, resume, delete, add).
