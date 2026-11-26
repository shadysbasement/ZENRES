-- ZENRES legal choices are append-only so the version accepted at registration remains auditable.
CREATE TABLE user_policy_events (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_id UUID,
    policy_id TEXT NOT NULL CHECK (policy_id IN ('terms', 'privacy_ack', 'marketing')),
    policy_version TEXT NOT NULL,
    accepted BOOLEAN NOT NULL,
    source TEXT NOT NULL CHECK (source IN ('registration', 'account_preferences')),
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX user_policy_events_user_time_idx ON user_policy_events (user_id, occurred_at DESC);
CREATE TRIGGER user_policy_events_immutable BEFORE UPDATE OR DELETE ON user_policy_events
    FOR EACH ROW EXECUTE FUNCTION forbid_audit_mutation();
