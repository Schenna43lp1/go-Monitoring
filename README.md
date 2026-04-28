# Minimal Go Agent

Kleiner HTTP-Agent, der JSON-Daten an ein Backend sendet.

## Start

```powershell
$env:BACKEND_URL="http://localhost:8080/ingest"
go run .
```

Der Agent läuft standardmäßig auf Port `9090`.

## Daten senden

```powershell
Invoke-RestMethod -Method Post http://localhost:9090/send `
  -ContentType "application/json" `
  -Body '{"agent_id":"agent-1","message":"hallo backend"}'
```

Oder Testpayload senden:

```powershell
Invoke-RestMethod http://localhost:9090/sample
```

## Konfiguration

- `BACKEND_URL`: Ziel-URL vom Backend, default `http://localhost:8080/ingest`
- `AGENT_PORT`: Port vom Agenten, default `9090`
- `AGENT_ID`: Agent-ID für `/sample`, default `agent-1`
