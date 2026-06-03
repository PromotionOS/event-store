-- +goose Up
CREATE TABLE IF NOT EXISTS domain_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    context VARCHAR(100) NOT NULL,
    aggregate_id VARCHAR(255),
    aggregate_type VARCHAR(100),
    event_type VARCHAR(100) NOT NULL,
    schema_version INTEGER NOT NULL DEFAULT 1,
    payload JSONB NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    recorded_at TIMESTAMPTZ DEFAULT NOW()
);

REVOKE UPDATE, DELETE ON domain_events FROM PUBLIC;

CREATE INDEX IF NOT EXISTS idx_events_tenant_context ON domain_events(tenant_id, context);
CREATE INDEX IF NOT EXISTS idx_events_aggregate ON domain_events(aggregate_id, aggregate_type);
CREATE INDEX IF NOT EXISTS idx_events_occurred ON domain_events(occurred_at);

-- +goose Down
DROP TABLE IF EXISTS domain_events;
