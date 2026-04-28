# Multi-Server Monitoring (Frontend + Backend + Agent)

This repository is a minimal example of a small monitoring platform consisting of three components:

- `backend/` — central Go API that receives metrics from agents and sends alerts (e.g., to Discord)
- `agent/` — lightweight Go agent that runs on hosts and sends CPU, RAM and disk metrics
- `frontend/` — static HTML/CSS/JS UI for metric visualization and alert-rule management

## Changelog

For full change history and release notes, see: [CHANGELOG](./changelog.md)

## Quick start

### 1) Start the backend locally

```powershell
cd backend
go run .
```

Default URL: `http://localhost:8080`

### 2) Run an agent on a host

Set `BACKEND_URL` and `AGENT_ID` for each agent. Example (PowerShell):

```powershell
cd agent
$env:BACKEND_URL = "http://localhost:8080/api/data"
$env:AGENT_ID = "server-1"
$env:AGENT_INTERVAL_SECONDS = "30"
go run .
```

Example for a second host:

```powershell
cd agent
$env:BACKEND_URL = "http://localhost:8080/api/data"
$env:AGENT_ID = "server-2"
go run .
```

### 3) Start the frontend

```powershell
cd frontend
python -m http.server 3000
```

Open the UI at `http://localhost:3000`

## API endpoints

- `GET /api/health` — Check backend health
- `GET /api/data` — Read stored metrics
- `POST /api/data` — Post metrics (agent -> backend)
- `GET /api/alerts` — Read alert configuration
- `POST /api/alerts` — Save alert configuration

## Discord alerts

The frontend allows configuring:

- Discord webhook URL
- Enable/disable alerts
- CPU, RAM and disk thresholds (percent)
- Cooldown (seconds) to avoid repeated notifications

When an agent exceeds a threshold, the backend will send a message to the configured Discord webhook.

## Agent configuration (env variables)

- `BACKEND_URL` — URL to the backend (default: `http://localhost:8080/api/data`)
- `AGENT_ID` — Unique ID per agent/host
- `AGENT_INTERVAL_SECONDS` — Send interval in seconds (default: `30`)

## Notes

- This project is a minimal example. For production use you should add authentication, secure storage of webhooks, rate limiting, tests and monitoring of agents.

If you want, I can add a short deployment guide (Docker/systemd) or example alert rules.
