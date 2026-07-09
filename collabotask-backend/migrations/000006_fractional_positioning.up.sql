-- Drop CHECK constraint that disallows negative positions (fractional positions
-- may be negative from head-inserts: new_head = first - STEP).
ALTER TABLE columns DROP CONSTRAINT IF EXISTS columns_position_check;

-- Change position columns from INTEGER to NUMERIC.
ALTER TABLE columns ALTER COLUMN position TYPE NUMERIC USING position::NUMERIC;
ALTER TABLE cards   ALTER COLUMN position TYPE NUMERIC USING position::NUMERIC;

-- Reseed existing rows to evenly-spaced STEP=1000 multiples so the initial gaps
-- are wide enough for many inserts before a rebalance is triggered. This makes
-- the migration correct against any database state, not just an empty one.
-- The literal 1000 mirrors domain.PositionStep (ADR-004); SQL cannot import the
-- Go const, so keep the two in sync if STEP ever changes.
UPDATE columns
SET position = rn.rn * 1000
FROM (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY board_id  ORDER BY position, id) AS rn
    FROM columns
) AS rn
WHERE columns.id = rn.id;

UPDATE cards
SET position = rn.rn * 1000
FROM (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY column_id ORDER BY position, id) AS rn
    FROM cards
) AS rn
WHERE cards.id = rn.id;
