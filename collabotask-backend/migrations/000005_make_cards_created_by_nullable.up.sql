-- cards.created_by used ON DELETE SET NULL on a NOT NULL column, which would
-- error when a referencing user is deleted. Make the column nullable so the
-- SET NULL policy is valid (a card outlives its creator's account).
ALTER TABLE cards ALTER COLUMN created_by DROP NOT NULL;
