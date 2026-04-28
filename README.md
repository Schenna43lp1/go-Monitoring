# Multi-Server Monitoring (Frontend + Backend + Agent)

Kurze Anleitung und Übersicht für das Beispielprojekt. Ziel ist ein leichtgewichtiges Monitoring-System mit mehreren Agenten, einem zentralen Go-Backend und einem einfachen Frontend zur Visualisierung.

## Repositorystruktur

- `backend/` — zentrale Go-API, empfängt Metriken von Agenten und versendet Alerts (z. B. an Discord)
- `agent/` — kleiner Go-Agent, läuft auf Hosts und sendet CPU-, RAM- und Disk-Metriken
- `frontend/` — statische HTML/CSS/JS-Oberfläche zur Ansicht von Metriken und Verwaltung von Alert-Regeln

## Changelog

Weitere Informationen zu Änderungen und Releases: [CHANGELOG](./changelog.md)

## Schnellstart

### 1) Backend lokal starten

```powershell
cd backend
go run .
```

Standard-URL: `http://localhost:8080`

### 2) Agent auf einem Host starten

Passen Sie `BACKEND_URL` und `AGENT_ID` für jeden Agent an. Beispiel (PowerShell):

```powershell
cd agent
$env:BACKEND_URL = "http://localhost:8080/api/data"
$env:AGENT_ID = "server-1"
$env:AGENT_INTERVAL_SECONDS = "30"
go run .
```

Beispiel für einen zweiten Host:

```powershell
cd agent
$env:BACKEND_URL = "http://localhost:8080/api/data"
$env:AGENT_ID = "server-2"
go run .
```

### 3) Frontend lokal starten

```powershell
cd frontend
python -m http.server 3000
```

Öffne das Frontend im Browser: `http://localhost:3000`

## API Endpoints

- `GET /api/health` — Prüfe den Backend-Status
- `GET /api/data` — Lese gespeicherte Metriken
- `POST /api/data` — Sende Metriken (Agent -> Backend)
- `GET /api/alerts` — Lese Alert-Konfiguration
- `POST /api/alerts` — Speichere Alert-Konfiguration

## Discord Alerts

Im Frontend können Nutzer folgende Einstellungen vornehmen:

- Discord Webhook URL
- Alert aktiv/inaktiv
- CPU-, RAM- und Disk-Grenzwerte (in Prozent)
- Cooldown (Sekunden) für wiederkehrende Alerts

Wenn ein Agent einen Grenzwert überschreitet, sendet das Backend eine Nachricht an den konfigurierten Discord Webhook.

## Agent Konfiguration (Umgebungsvariablen)

- `BACKEND_URL` — URL zum Backend (Standard: `http://localhost:8080/api/data`)
- `AGENT_ID` — Eindeutige ID pro Agent / Host
- `AGENT_INTERVAL_SECONDS` — Sendeintervall in Sekunden (Standard: `30`)

## Weitere Hinweise

- Dieses Projekt ist ein Minimalbeispiel. Für Produktion wären zusätzliche Schritte nötig: Authentifizierung, sichere Speicherung von Webhooks, Rate limiting, Tests und Monitoring der Agenten.

Wenn du möchtest, schreibe ich einen kurzen Abschnitt zur Deployment-Strategie (Docker / systemd) oder füge Beispiele für Alert-Regeln hinzu.
