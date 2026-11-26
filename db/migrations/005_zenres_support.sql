-- ZENRES support records. Customer reads are scoped to the requester in the API.
CREATE TABLE support_tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_no BIGINT GENERATED ALWAYS AS IDENTITY UNIQUE,
    requester_user_id UUID NOT NULL REFERENCES users(id),
    org_id UUID, -- optional until the separate core-platform schema is deployed here
    plan_at_creation TEXT NOT NULL DEFAULT 'unassigned',
    kind TEXT NOT NULL CHECK (kind IN ('question','live_chat','callback','bug','feedback')),
    category TEXT NOT NULL CHECK (category IN ('platform','tally','cbam','trace','freight','marketplace','billing','security','other')),
    priority TEXT NOT NULL CHECK (priority IN ('low','normal','high','urgent')),
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','pending','escalated','resolved')),
    subject TEXT NOT NULL CHECK (length(subject) BETWEEN 5 AND 160),
    assigned_admin_key_id UUID REFERENCES admin_pgp_keys(id),
    callback_at TIMESTAMPTZ,
    callback_timezone TEXT,
    financial_access_granted BOOLEAN NOT NULL DEFAULT FALSE,
    first_response_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX support_tickets_requester_idx ON support_tickets (requester_user_id, created_at DESC);
CREATE INDEX support_tickets_queue_idx ON support_tickets (status, priority, created_at);

CREATE TABLE support_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    author_user_id UUID REFERENCES users(id),
    author_admin_key_id UUID REFERENCES admin_pgp_keys(id),
    body TEXT NOT NULL CHECK (length(body) BETWEEN 1 AND 4000),
    internal_note BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK ((author_user_id IS NOT NULL) <> (author_admin_key_id IS NOT NULL)),
    CHECK (NOT internal_note OR author_admin_key_id IS NOT NULL)
);
CREATE INDEX support_messages_ticket_idx ON support_messages (ticket_id, created_at);

CREATE TABLE support_feedback (
    ticket_id UUID PRIMARY KEY REFERENCES support_tickets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT NOT NULL DEFAULT '' CHECK (length(comment) <= 1000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE support_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    ticket_id UUID REFERENCES support_tickets(id),
    actor_user_id UUID REFERENCES users(id),
    actor_admin_key_id UUID REFERENCES admin_pgp_keys(id),
    action TEXT NOT NULL,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX support_audit_ticket_idx ON support_audit_logs (ticket_id, created_at);
CREATE TRIGGER support_audit_immutable BEFORE UPDATE OR DELETE ON support_audit_logs
    FOR EACH ROW EXECUTE FUNCTION forbid_audit_mutation();

CREATE TABLE support_announcements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL CHECK (length(title) BETWEEN 5 AND 160),
    body TEXT NOT NULL CHECK (length(body) BETWEEN 5 AND 2000),
    state TEXT NOT NULL CHECK (state IN ('maintenance','incident','resolved')),
    published_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ends_at TIMESTAMPTZ,
    created_by UUID NOT NULL REFERENCES admin_pgp_keys(id)
);
CREATE INDEX support_announcements_public_idx ON support_announcements (published_at DESC);
