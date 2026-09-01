CREATE TABLE family_links (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    child_id   UUID NOT NULL,
    parent_id  UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_family_links_parent_id ON family_links (parent_id);
CREATE INDEX idx_family_links_child_id ON family_links (child_id);
