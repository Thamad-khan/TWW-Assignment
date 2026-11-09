# Campaign Events API

Go HTTP service for receiving provider events and returning campaign statistics.

## Requirements
Go 1.22 or later.

## Run
cd starter
go run .

## Endpoints

POST /events

Content-Type: application/json

Send a JSON array of events.

GET /campaigns/{campaign_id}/stats

GET /health

## Example

curl -X POST localhost:8080/events \
-H 'Content-Type: application/json' \
-d '[{"event_id":"evt_1","campaign_id":"cmp_1","contact_id":"ct_1","type":"delivered","timestamp":"2026-08-10T06:15:00Z"}]'

curl localhost:8080/campaigns/cmp_1/stats

## Design decisions
- Event ID is the global idempotency key.
- Duplicate event IDs are ignored.
- Valid items in a partially invalid batch are still accepted.
- Unique opens count distinct contacts per campaign.
- State is protected by a mutex.
- Storage is in-memory and resets on process restart.
- Request body is limited to 2 MiB.
