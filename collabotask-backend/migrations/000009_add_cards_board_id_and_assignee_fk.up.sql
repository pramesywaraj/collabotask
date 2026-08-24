-- 1. add the denormalized column, nullable for backfill
ALTER TABLE cards ADD COLUMN board_id UUID;

-- 2. backfill from each card's column's board
UPDATE cards c SET board_id = col.board_id
FROM columns col WHERE c.column_id = col.id;

-- 3. lock it down
ALTER TABLE cards ALTER COLUMN board_id SET NOT NULL;
ALTER TABLE cards
  ADD CONSTRAINT fk_cards_board FOREIGN KEY (board_id)
  REFERENCES boards(id) ON DELETE CASCADE;

-- 4. null any assignee that is not a member of the card's board, before adding
--    the composite FK. On a clean database this is a no-op.
--    Data note: nulled assignments are not restorable (they were invalid).
DO $$
DECLARE n int;
BEGIN
  UPDATE cards c SET assigned_to = NULL
  WHERE c.assigned_to IS NOT NULL
    AND NOT EXISTS (
      SELECT 1 FROM board_members bm
      WHERE bm.board_id = c.board_id AND bm.user_id = c.assigned_to
    );
  GET DIAGNOSTICS n = ROW_COUNT;
  RAISE NOTICE 'nulled % violating assignees', n;
END $$;

-- 5. the invariant: a card's assignee must be a board member of the card's board
ALTER TABLE cards
  ADD CONSTRAINT fk_cards_assignee_board_member
  FOREIGN KEY (board_id, assigned_to)
  REFERENCES board_members (board_id, user_id)
  ON DELETE SET NULL (assigned_to);
