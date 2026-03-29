# Revenge — External API Inventory

Complete list of all external APIs, services, and integrations used in the Revenge media server.

---

## 1. Metadata Providers (12)

All providers live in `internal/service/metadata/providers/` with full HTTP client implementations, rate limiting, caching, and circuit breakers.

| # | Provider | Purpose | Auth | Key Path |
|---|----------|---------|------|----------|
| 1 | **TMDb** (The Movie Database) | Primary movie/TV/person/image/collection metadata | API key | `internal/service/metadata/providers/tmdb/` |
| 2 | **TheTVDB** | TV metadata, episode data | API key | `internal/service/metadata/providers/tvdb/` |
| 3 | **Fanart.tv** | High-quality artwork (posters, backgrounds, logos) | API key + optional client key | `internal/service/metadata/providers/fanarttv/` |
| 4 | **OMDb** | Ratings aggregation (IMDb, Rotten Tomatoes, Metacritic) | API key | `internal/service/metadata/providers/omdb/` |
| 5 | **TVmaze** | Free TV metadata | None (public) | `internal/service/metadata/providers/tvmaze/` |
| 6 | **AniList** | Anime metadata (GraphQL) | None (public) | `internal/service/metadata/providers/anilist/` |
| 7 | **Kitsu** | Anime/manga metadata (JSON:API) | None (public) | `internal/service/metadata/providers/kitsu/` |
| 8 | **AniDB** | Detailed anime episode data, characters, tags | Client name + version | `internal/service/metadata/providers/anidb/` |
| 9 | **MyAnimeList (MAL)** | Anime ratings, rankings, community data | Client ID | `internal/service/metadata/providers/mal/` |
| 10 | **Trakt** | Movie/TV community metadata, ratings, cross-referenced IDs | Client ID | `internal/service/metadata/providers/trakt/` |
| 11 | **Simkl** | Movies, TV, anime tracking with cross-referenced IDs | Client ID | `internal/service/metadata/providers/simkl/` |
| 12 | **Letterboxd** | Movie-focused community ratings and reviews | API key + API secret | `internal/service/metadata/providers/letterboxd/` |

---

## 2. *arr Integrations (2)

Bidirectional library sync with webhook event handling. Located in `internal/integration/`.

| # | Service | Purpose | Auth | Key Path |
|---|---------|---------|------|----------|
| 1 | **Radarr** | Movie library management, webhook sync | Base URL + API key + webhook secret | `internal/integration/radarr/` |
| 2 | **Sonarr** | TV show library management, webhook sync | Base URL + API key + webhook secret | `internal/integration/sonarr/` |

Features: auto-sync with configurable intervals, quality profile management, health checks.

---

## 3. Notification Services (5)

Centralized notification hub in `internal/service/notification/agents/`.

| # | Service | Purpose | Auth | Key Path |
|---|---------|---------|------|----------|
| 1 | **Discord** | Event notifications via webhooks | Webhook URL | `internal/service/notification/agents/discord.go` |
| 2 | **Gotify** | Self-hosted push notifications | Server URL + app token | `internal/service/notification/agents/gotify.go` |
| 3 | **ntfy** | Simple push notifications (ntfy.sh compatible) | Server URL + topic + optional token | `internal/service/notification/agents/` |
| 4 | **Generic Webhook** | Flexible HTTP webhooks (POST/PUT/PATCH) | URL + optional headers/auth | `internal/service/notification/agents/webhook.go` |
| 5 | **Email** | Email-based alerts | Uses email service (see below) | `internal/service/notification/agents/email.go` |

---