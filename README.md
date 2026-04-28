# Multi-Server Agent mit Frontend und Backend

Minimalbeispiel mit drei getrennten Teilen.

- `backend/`: zentrale Go-API, nimmt Daten von allen Agenten an
- `agent/`: kleiner Go-Agent, laeuft auf mehreren Servern und sendet Heartbeats
- `frontend/`: statische HTML/CSS/JS-Oberflaeche, liest Daten vom Backend

## 1. Backend zentral starten

```powershell
cd backend
go run .
```

Das Backend laeuft standardmaessig auf `http://localhost:8080`.

## 2. Agent auf jedem Server starten

Auf jedem Server dieselbe Agent-App starten, aber mit eigener `AGENT_ID`.

```powershell
cd agent
$env:BACKEND_URL="http://DEIN-BACKEND-SERVER:8080/api/data"
$env:AGENT_ID="server-1"
$env:AGENT_INTERVAL_SECONDS="30"
go run .
```

Beispiel fuer einen zweiten Server:

```powershell
cd agent
$env:BACKEND_URL="http://DEIN-BACKEND-SERVER:8080/api/data"
$env:AGENT_ID="server-2"
go run .
```

## 3. Frontend starten

```powershell
cd frontend
python -m http.server 3000
```

Dann im Browser oeffnen:

`http://localhost:3000`

## API

- `GET /api/health`: Status pruefen
- `GET /api/data`: gespeicherte Daten lesen
- `POST /api/data`: Daten speichern

## Agent-Konfiguration

- `BACKEND_URL`: zentrale Backend-URL, default `http://localhost:8080/api/data`
- `AGENT_ID`: eindeutiger Name pro Server, default basiert auf Hostname
- `AGENT_INTERVAL_SECONDS`: Sendeintervall, default `30`
