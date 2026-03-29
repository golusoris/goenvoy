# Lurkarr — External API Integrations

Vollständige Aufstellung aller externen APIs, die von Lurkarr unterstützt oder genutzt werden.

---

## 1. Arr Stack Clients

Paket: `internal/arrclient/`

| Service | API-Version | Auth | Zweck |
|---------|-------------|------|-------|
| **Sonarr** | v3 | X-Api-Key Header | TV-Serien-Management, Suche, Queue-Monitoring |
| **Radarr** | v3 | X-Api-Key Header | Film-Management, Suche, Queue-Monitoring |
| **Lidarr** | v1 | X-Api-Key Header | Musik-/Album-Management |
| **Readarr** | v1 | X-Api-Key Header | Buch-Management |
| **Whisparr v2** | v3 | X-Api-Key Header | Adult Content (Sonarr-basiert: Studios=Series, Scenes=Episodes) |
| **Whisparr v3 (Eros)** | v3 | X-Api-Key Header | Adult Content (Radarr-basiert: Einzel-Items) |
| **Prowlarr** | v1 | X-Api-Key Header | Indexer-Management & Statistiken |

**Gemeinsame Endpunkte:**
- `GET /api/{v}/wanted/missing` — fehlende Items
- `GET /api/{v}/wanted/cutoff` — upgrade-fähige Items
- `GET /api/{v}/queue` — Download-Queue
- `POST /api/{v}/command` — Aktionen auslösen
- `DELETE /api/{v}/queue/{id}` — Queue-Items entfernen
- `GET /api/{v}/system/status` — Versionsinformationen

---

## 2. Subtitle-Management

Paket: `internal/bazarrclient/`

| Service | Auth | Zweck |
|---------|------|-------|
| **Bazarr** | X-API-Key Header | Untertitel-Management (fehlende Untertitel, Provider-Health, Download-History) |

**API-Basis:** `/api/`

---

## 3. Comic-/Manga-Management

Paket: `internal/kapowarrclient/`

| Service | Auth | Zweck |
|---------|------|-------|
| **Kapowarr** | api_key Query-Parameter | Comic-/Manga-Management (Volume-Stats, Queue, Collection-Updates) |

**API-Basis:** `/api/`

---

## 4. Anime-Management

Paket: `internal/shokoclient/`

| Service | Auth | Zweck |
|---------|------|-------|
| **Shoko Server** | apikey Header | Anime-Verwaltung (Collection-Stats, Serien-Übersicht, Episode-Tracking) |

**API-Basis:** `/api/v3`

---

## 5. Request-Management

Paket: `internal/seerr/`

| Service | Auth | Zweck |
|---------|------|-------|
| **Seerr** | X-Api-Key Header | Media-Request-Verwaltung (Requests, Approval/Decline, Auto-Requests, Status) |

**API-Basis:** `/api/v1`

---

## 6. Download-Clients

Paket: `internal/downloadclients/`

### Torrent-Clients

| Client | Protokoll | Auth | Zweck |
|--------|-----------|------|-------|
| **qBittorrent** | WebUI API v2 | Cookie-basiert | Torrent-Downloads |
| **Transmission** | JSON-RPC 2.0 | HTTP Basic / Session-Header | Torrent-Downloads |
| **Deluge** | JSON-RPC | Cookie-basiert | Torrent-Downloads |
| **rTorrent** | XML-RPC (go-rtorrent) | HTTP Basic | Torrent-Downloads |
| **uTorrent** | WebUI API | Cookie-basiert | Torrent-Downloads |

### Usenet-Clients

| Client | Protokoll | Auth | Zweck |
|--------|-----------|------|-------|
| **SABnzbd** | REST + XML | API Key | Usenet-Downloads |
| **NZBGet** | XML-RPC | HTTP Basic | Usenet-Downloads |

**Gemeinsame Operationen:** Queue/History abrufen, Pause/Resume, Löschen, Recheck, Status

---

## 7. Notification-Services

Paket: `internal/notifications/`

| Provider | Typ | Endpunkt | Auth |
|----------|-----|----------|------|
| **Discord** | Webhook | User-konfigurierte Webhook-URL | Token in URL |
| **Telegram** | Bot API | `api.telegram.org/bot{token}/sendMessage` | Bot-Token |
| **Gotify** | Push | `{ServerURL}/message` | X-Gotify-Key Header |
| **ntfy** | Simple HTTP | `{ServerURL}/{topic}` | Optional Bearer-Token |
| **Pushover** | REST API | `api.pushover.net/1/messages.json` | Token + User-Key |
| **Email** | SMTP/SMTPS | User-konfiguriert | Optional (Plain Auth) |
| **Apprise** | REST | `{ServerURL}/notify/` | Keine (self-hosted) |
| **Webhook** | JSON POST | User-konfiguriert | Custom Headers |

**Event-Typen:** Lurk started/completed, Queue-Removal, Download stuck, Scheduler-Action, Errors, Test

---

## 8. Authentifizierung

Paket: `internal/auth/`

| Methode | Provider | Zweck |
|---------|----------|-------|
| **OIDC** | Keycloak, Okta, Authentik, AzureAD, etc. | Enterprise SSO, Federated Identity |
| **WebAuthn** | W3C Standard (FIDO2/Passkeys) | Passwortlose Authentifizierung |
| **Proxy Auth** | Authelia, Authentik, Nginx, etc. | Forward-Proxy-Integration |
| **Local** | Built-in | Username/Password mit bcrypt + TOTP 2FA |

---

## 9. Monitoring & Observability

### Prometheus Metrics

Paket: `internal/metrics/`

- **Endpunkt:** `/metrics`
- **Metriken:** Lurk-Ops, Queue-Cleaner, Download-Clients, Scheduler, HTTP-Requests, External-API-Calls

### Datenbank

| Service | Treiber | Zweck |
|---------|---------|-------|
| **PostgreSQL** | pgx/v5 | Primäre Datenbank (sole database, siehe ADR-005) |

### Deployment-Integration

- **Grafana** — Dashboards & Provisioning (in `deploy/grafana/`)
- **Loki** — Log-Aggregation (in `deploy/loki.yml`)

---

## Zusammenfassung

| Kategorie | Anzahl | Services |
|-----------|--------|----------|
| Arr Stack | 7 | Sonarr, Radarr, Lidarr, Readarr, Whisparr v2, Whisparr v3 (Eros), Prowlarr |
| Content-Manager | 3 | Bazarr, Kapowarr, Shoko |
| Request-Manager | 1 | Seerr |
| Download-Clients | 7 | qBittorrent, Transmission, Deluge, rTorrent, uTorrent, SABnzbd, NZBGet |
| Notifications | 8 | Discord, Telegram, Gotify, ntfy, Pushover, Email, Apprise, Webhook |
| Auth-Provider | 4 | OIDC, WebAuthn, Proxy Auth, Local |
| Monitoring/Infra | 3 | Prometheus, PostgreSQL, Grafana/Loki |

**Gesamt: 33 externe Service-Integrationen**
