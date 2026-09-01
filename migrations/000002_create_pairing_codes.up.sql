CREATE TABLE pairing_codes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code       CHAR(6) NOT NULL,
    parent_id  UUID NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- At most one *active* (unused) code may share a value at any moment.
CREATE UNIQUE INDEX idx_pairing_codes_active_code ON pairing_codes (code) WHERE used_at IS NULL;
CREATE INDEX idx_pairing_codes_parent_id ON pairing_codes (parent_id);
