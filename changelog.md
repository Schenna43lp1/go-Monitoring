# Changelog

All notable changes to this project are recorded here in chronological order. This document follows a simple "Keep a Changelog" style (simplified).

## Unreleased

### Added (Unreleased)

- Go monitoring agent: Minimal Go-based agent to collect and post metrics to the backend.
- Discord integration: Support for Discord webhooks to deliver real-time monitoring alerts.

### Changed (Unreleased)

- Project refactor: Repository restructured into a modular layout by separating Backend, Agent, and Frontend components.
- Decoupled logic: Implemented dedicated responsibilities for the agent and backend to improve API handling and scalability.
- Frontend interface improvements: Enhanced UI for better data visualization and interaction.
- Improved API handling: Optimized backend processing of incoming agent data for greater reliability.

### Fixed (Unreleased)

- System metrics alerts: Configured alert thresholds for critical system metrics (planned/ongoing).

---

## [0.1.0] - 2026-04-28

### Added (0.1.0)

- Initial project skeleton with separate components:
  - `agent/` (Go agent to collect and send metrics)
  - `backend/` (Go backend for processing and storage)
  - `frontend/` (simple UI for visualization)

- README, base structure and initial configuration
- Minimal monitoring agent written in Go
- Discord webhook integration for alerts

### Changed (0.1.0)

- Refactored project structure to decouple agent, backend and frontend.

### Maintenance (0.1.0)

- Added a `.gitignore` for Go projects

- Removed accidentally committed build artifacts (e.g. `.gocache`)

## [0.1.1] - 2026-04-28

### Changed (0.1.1)

- README, verbessert
