-- Revert to NOT NULL. Note: fails if any cards.created_by are NULL.
ALTER TABLE cards ALTER COLUMN created_by SET NOT NULL;
