CREATE TABLE activities (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    board_id    UUID NOT NULL REFERENCES boards(id)  ON DELETE CASCADE,
    user_id     UUID          REFERENCES users(id)   ON DELETE SET NULL,
    action_type VARCHAR NOT NULL,
    entity_type VARCHAR NOT NULL,
    entity_id   UUID NOT NULL,
    metadata    JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_activities_board_created ON activities (board_id, created_at DESC);
