# Event Store

Pure infrastructure — consumes ALL domain events and appends to append-only PostgreSQL.
No business logic. Pre-built by facilitator. Teams do not modify this service.

## Purpose
- Audit trail for all domain events
- AI/ML readiness — versioned event history

## Channels Subscribed
- promotionos.campaign.published
- promotionos.campaign.paused
- promotionos.catalog.item.excluded
- promotionos.customer.segment.updated
- promotionos.redemption.redeemed
- promotionos.redemption.claim.submitted
- promotionos.analytics.budget.exhausted
- promotionos.campaign.budget.updated

## Rules
- APPEND ONLY. No UPDATE. No DELETE. Ever.
- Teams can READ from this for debugging during the session.

## API
- `GET /health` — health check
- `GET /events?tenantId=X&aggregateId=Y` — events by aggregate
- `GET /events?tenantId=X&context=campaign&from=2026-06-01&to=2026-06-30` — events by context

## Local Development
```bash
export DB_URL=postgresql://localhost:5432/railway
export REDIS_URL=redis://localhost:6379
go run cmd/main.go
```
