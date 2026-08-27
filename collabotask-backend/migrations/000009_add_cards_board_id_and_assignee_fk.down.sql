-- data loss: assignees nulled during the up migration are not restorable.
ALTER TABLE cards DROP CONSTRAINT IF EXISTS fk_cards_assignee_board_member;
ALTER TABLE cards DROP CONSTRAINT IF EXISTS fk_cards_board;
ALTER TABLE cards DROP COLUMN IF EXISTS board_id;
