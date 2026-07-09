-- Reseed to contiguous 0-based integers before changing the type back.
-- Truncation alone would be lossy (fractions collide); bundling the re-seed
-- makes the rollback deterministic and self-healing.
UPDATE columns
SET position = rn.rn - 1
FROM (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY board_id  ORDER BY position, id) AS rn
    FROM columns
) AS rn
WHERE columns.id = rn.id;

UPDATE cards
SET position = rn.rn - 1
FROM (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY column_id ORDER BY position, id) AS rn
    FROM cards
) AS rn
WHERE cards.id = rn.id;

-- Change back to INTEGER (order-preserving after the re-seed above).
ALTER TABLE cards   ALTER COLUMN position TYPE INTEGER USING position::INTEGER;
ALTER TABLE columns ALTER COLUMN position TYPE INTEGER USING position::INTEGER;

-- Restore the non-negative CHECK constraint.
ALTER TABLE columns ADD CONSTRAINT columns_position_check CHECK (position >= 0);
