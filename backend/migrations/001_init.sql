CREATE TABLE IF NOT EXISTS meters (
    meter_id TEXT PRIMARY KEY,
    provider TEXT,
    region TEXT
);

CREATE TABLE IF NOT EXISTS tariffs (
    tariff_id SERIAL PRIMARY KEY,
    provider TEXT NOT NULL,
    region TEXT NOT NULL DEFAULT '',
    rate_type TEXT NOT NULL DEFAULT 'industrial',
    price_per_kwh DOUBLE PRECISION NOT NULL CHECK (price_per_kwh > 0),
    currency TEXT NOT NULL DEFAULT 'COP',
    valid_from TIMESTAMPTZ NOT NULL,
    valid_to TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS billing (
    billing_id SERIAL PRIMARY KEY,
    meter_id TEXT NOT NULL REFERENCES meters(meter_id) ON DELETE CASCADE,
    tariff_id INTEGER REFERENCES tariffs(tariff_id) ON DELETE SET NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    total_kwh DOUBLE PRECISION NOT NULL DEFAULT 0,
    price_per_kwh DOUBLE PRECISION NOT NULL CHECK (price_per_kwh > 0),
    total_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'COP',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS readings (
    meter_id TEXT NOT NULL REFERENCES meters(meter_id) ON DELETE CASCADE,
    recorded_at TIMESTAMPTZ NOT NULL,
    consumption_kwh DOUBLE PRECISION NOT NULL,
    voltage_v DOUBLE PRECISION NOT NULL,
    current_a DOUBLE PRECISION NOT NULL,
    power_factor DOUBLE PRECISION NOT NULL,
    status TEXT NOT NULL,
    PRIMARY KEY (meter_id, recorded_at)
);

CREATE INDEX IF NOT EXISTS readings_recorded_at_idx ON readings(recorded_at);

CREATE TABLE IF NOT EXISTS events (
    meter_id TEXT NOT NULL REFERENCES meters(meter_id) ON DELETE CASCADE,
    occurred_at TIMESTAMPTZ NOT NULL,
    event_type TEXT NOT NULL,
    description TEXT NOT NULL,
    PRIMARY KEY (meter_id, occurred_at, event_type)
);

CREATE TABLE IF NOT EXISTS analyses (
    id UUID PRIMARY KEY,
    status TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    anomaly_count INTEGER NOT NULL DEFAULT 0,
    anomalies JSONB NOT NULL DEFAULT '[]'::jsonb,
    error TEXT NOT NULL DEFAULT '',
    ai_response JSONB,
    ai_error TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS ai_insights (
    meter_id TEXT NOT NULL,
    from_ts TIMESTAMPTZ NOT NULL,
    to_ts TIMESTAMPTZ NOT NULL,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    payload JSONB NOT NULL,
    PRIMARY KEY (meter_id, from_ts, to_ts)
);

ALTER TABLE analyses ADD COLUMN IF NOT EXISTS ai_response JSONB;
ALTER TABLE analyses ADD COLUMN IF NOT EXISTS ai_error TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS anomalies (
    id TEXT PRIMARY KEY,
    analysis_id UUID NOT NULL REFERENCES analyses(id) ON DELETE CASCADE,
    meter_id TEXT NOT NULL REFERENCES meters(meter_id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    severity TEXT NOT NULL,
    confidence DOUBLE PRECISION NOT NULL,
    reason TEXT NOT NULL,
    recommended_action TEXT NOT NULL,
    detected_at TIMESTAMPTZ NOT NULL,
    evidence JSONB NOT NULL
);