-- Single-use admin handoff between the public site and app, shared across auth instances.
CREATE TABLE admin_handoff_codes (
    code_hash     BYTEA PRIMARY KEY,
    session_token TEXT NOT NULL,
    expires_at    TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON admin_handoff_codes (expires_at);
