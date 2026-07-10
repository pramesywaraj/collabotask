-- Board visibility (SRS §2.2, ADR-005). WORKSPACE (default) or PRIVATE.
ALTER TABLE boards
    ADD COLUMN visibility VARCHAR(20) NOT NULL DEFAULT 'WORKSPACE'
        CHECK (visibility IN ('PRIVATE', 'WORKSPACE'));

-- Schema hygiene / forward-compat only: created_by can never be NULL in Phase 1
-- (no account deletion), but demoting it to nullable + ON DELETE SET NULL lets a
-- future account-deletion feature (Phase 2, SRS §10) remove a user without
-- deleting their boards. board_members remains the sole source of ownership.
ALTER TABLE boards ALTER COLUMN created_by DROP NOT NULL;
ALTER TABLE boards DROP CONSTRAINT boards_created_by_fkey;
ALTER TABLE boards
    ADD CONSTRAINT boards_created_by_fkey
        FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
