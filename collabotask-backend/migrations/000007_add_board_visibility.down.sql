ALTER TABLE boards DROP CONSTRAINT boards_created_by_fkey;
ALTER TABLE boards
    ADD CONSTRAINT boards_created_by_fkey
        FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT;
ALTER TABLE boards ALTER COLUMN created_by SET NOT NULL;

ALTER TABLE boards DROP COLUMN visibility;
