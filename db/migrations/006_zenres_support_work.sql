-- Actual support effort and costs, recorded by authorised agents; no inferred staffing rate.
CREATE TABLE support_work_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    admin_key_id UUID NOT NULL REFERENCES admin_pgp_keys(id),
    minutes INTEGER NOT NULL CHECK (minutes BETWEEN 1 AND 480),
    cost_amount NUMERIC(12,2),
    cost_currency CHAR(3) CHECK (cost_currency IN ('USD','INR','EUR')),
    note TEXT NOT NULL DEFAULT '' CHECK (length(note) <= 500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK ((cost_amount IS NULL AND cost_currency IS NULL) OR (cost_amount >= 0 AND cost_currency IS NOT NULL))
);
CREATE INDEX support_work_logs_ticket_idx ON support_work_logs (ticket_id, created_at);
CREATE TRIGGER support_work_logs_immutable BEFORE UPDATE OR DELETE ON support_work_logs
    FOR EACH ROW EXECUTE FUNCTION forbid_audit_mutation();
