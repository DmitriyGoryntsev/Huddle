CREATE TABLE auth.outbox (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type    VARCHAR(100) NOT NULL,
    payload       JSONB NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at  TIMESTAMPTZ,
    attempts      INTEGER DEFAULT 0,
    status        VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'published', 'failed'))
);

CREATE INDEX idx_outbox_pending ON auth.outbox(status, published_at) WHERE status = 'pending';
CREATE INDEX idx_outbox_created_at ON auth.outbox(created_at);