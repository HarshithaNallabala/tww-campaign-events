# Campaign Event Processing Service

## Overview

This project implements a Go-based backend service for ingesting campaign events and retrieving campaign statistics.

The service supports:
- Batch event ingestion through `POST /events`
- Campaign statistics through `GET /campaigns/{campaign_id}/stats`
- Event validation
- Duplicate event ID detection
- JSON responses
- Basic concurrent access protection

## Tech Stack

- Go
- `net/http`
- `encoding/json`
- `sync`
- Go testing package
- SQLite was not required for this implementation; events are stored in memory.

## Project Structure

```text
TWW Main Task/
├── NOTES.md
├── BUGS.md
├── AI_USAGE.md
├── README.md
├── starter/
│   ├── main.go
│   ├── types.go
│   ├── main_test.go
│   ├── go.mod
│   └── seed/
│       └── events.json
└── debugging/
    ├── main.go
    ├── events.jsonl
    └── expected_output.txt
```

## Running the Service

Prerequisite: Go installed.

Open a terminal and run:

```bash
cd starter
go run .
```

The service starts on:

```text
http://localhost:8080
```

## API Endpoints

### 1. Ingest Events

**Endpoint:** `POST /events`

Accepts a JSON array of campaign events.

Example request:

```json
[
  {
    "event_id": "evt_001",
    "campaign_id": "cmp_001",
    "contact_id": "contact_001",
    "type": "sent",
    "timestamp": "2026-10-30T10:00:00Z"
  },
  {
    "event_id": "evt_002",
    "campaign_id": "cmp_001",
    "contact_id": "contact_001",
    "type": "opened",
    "timestamp": "2026-10-30T10:05:00Z"
  }
]
```

Example PowerShell request:

```powershell
$body = Get-Content .\seed\events.json -Raw

Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8080/events `
  -ContentType "application/json" `
  -Body $body
```

The response contains accepted, duplicate, and rejected event counts.

### 2. Get Campaign Statistics

**Endpoint:** `GET /campaigns/{campaign_id}/stats`

Example:

```text
GET http://localhost:8080/campaigns/cmp_001/stats
```

Example PowerShell:

```powershell
Invoke-RestMethod http://localhost:8080/campaigns/cmp_001/stats
```

The response contains campaign event statistics, including sent, delivered, opened, and clicked counts.

## Running Tests

From the `starter` directory:

```bash
go test ./...
```

To run tests with verbose output:

```bash
go test -v ./...
```

## Debugging Program

From the `debugging` directory:

```bash
go run . events.jsonl
```

The program processes the event stream and prints campaign statistics.

The debugging fixes and their explanations are documented in `BUGS.md`.

## Current Limitations

- Events are stored in memory and are lost when the service restarts.
- No authentication or authorization is implemented.
- No persistent database is used.
- The service is intended as an assignment implementation, not a production deployment.
