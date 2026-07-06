# ADR-004: Positioning Strategy — Fractional NUMERIC + Repository-Layer Rebalancing

- **Status:** Accepted
- **Date:** 2026-07-06
- **Scope:** How card and column order is stored and mutated in `collabotask-backend` (UC-13d column reorder, UC-15 card move). Supersedes the integer-shift implementation flagged in SRS §9 (P1). Aligns with SRS §3.2.

## Context

Order for cards (within a column) and columns (within a board) was stored as `INTEGER` and mutated by **shifting neighbors**:

- **Column reorder** (`update_column_position.go`) took a 0-based target *index*, then renumbered **every** column in the board `0,1,2,…` in one write.
- **Card move** (`card.go` `Move`) took a 0-based target *index*, locked the card `FOR UPDATE`, and ran two UPDATEs that shifted every neighbor at or beyond the vacated slot (source column) and at or beyond the target slot (destination column).

**The root property both share: a single logical move rewrites many rows, not one.** To move one item, integer-shift must re-number its neighbors to keep positions contiguous — a column reorder rewrites the board's *entire* column set; a card move rewrites a whole range of cards, up to most of both columns. This write amplification is what makes both problems below possible — with a strategy where a move touches exactly one row, neither problem can occur.

Two problems make this the wrong foundation for Phase 1:

1. **Column reorder loses updates under concurrency.** Two simultaneous reorders both read the same snapshot, each computes a full renumber, and the later commit overwrites the earlier one wholesale. The losing user's move silently disappears — no error, a board state neither user intended. This is a correctness bug, not just a performance concern.
2. **Card move doesn't fit the realtime contract.** The `FOR UPDATE` transaction serializes concurrent drags, and — more decisively — a single move changes *many* rows. The WebSocket layer (next build step) defines `CARD_MOVED`/`COLUMN_MOVED` with a **single** `position` field (SRS §5.2). Integer-shift would force either N broadcast events per move (clients must apply all in order or diverge) or a blunt whole-column resync.

**Sequencing:** this migration lands *before* the WebSocket layer, so the realtime broadcast is built on single-row moves from the start rather than retrofitted.

## Options Considered

### Positioning scheme

- **Integer-shift (current, rejected)** — simple mental model, but loses updates (column reorder) and produces multi-row changes that break the single-`position` broadcast contract.
- **Lexorank (rejected)** — string-rank scheme used by Jira. No precision-exhaustion math, but overkill for Phase 1 scale and heavier to reason about. Kept as a Phase-2 escape hatch (SRS §8) if scale ever demands it.
- **Fractional NUMERIC (chosen)** — order is a `NUMERIC` coordinate. Insert between A and B = `(A+B)/2`; insert at ends = `first−STEP` / `last+STEP`. One UPDATE per move, no neighbor touched.

### Numeric Go type — `float64` vs `shopspring/decimal`

Chose **`float64`**. Rebalancing (below) fires when a neighbor gap drops under `0.5` — roughly 11 consecutive bisections of the same gap. `float64` only loses ordering precision after ~50 bisections of the same gap, twelve orders of magnitude below the threshold. The failure scenario `shopspring/decimal` guards against **cannot be reached** because rebalancing intervenes first. Positions are not financial values; exact decimal representation buys nothing here, and `decimal` would add a dependency plus a custom pgx codec for `NUMERIC`.

### Who computes the midpoint — client vs server

Chose **client**. The client already holds neighbor positions in its rendered board state and (with dnd-kit) knows the drop neighbors, so it computes `(A+B)/2` for free and sends the resulting value as `to_position`. The server trusts it and does one UPDATE — no extra read to re-derive neighbors. The stale-state risk (two clients computing the same midpoint before a `CARD_MOVED` arrives) produces at worst a transient position *tie*, not data loss, and the WebSocket layer keeps the window to tens of milliseconds. Server-side computation would trade an extra DB read and more server logic to guard against a tie that the deterministic ordering (below) already resolves.

### Where rebalancing lives — repository (Option A) vs use-case-owned transaction (Option B)

Chose **Option A: the repository**. Rebalancing must be **atomic with the move** (same transaction — no window where a position is ambiguous), and this codebase already manages the move transaction inside the repository method. Option B (use case owns the transaction, calls repo primitives) would require threading a `pgx.Tx`/unit-of-work through every repository method touched — a cross-cutting refactor unrelated to positioning.

The deciding reframe: **rebalancing is storage mechanics, not domain logic.** It exists *only* because a finite-precision numeric representation runs out of midpoints — swap to Lexorank or a linked list and the threshold vanishes. A rule that exists solely because of a storage choice belongs next to the SQL, not in the use case. Keeping it in the repository also means the use case never learns that positions are `float64` or that `STEP=1000`; if positioning is ever swapped (§8), only the repository changes.

## Decision

**Fractional NUMERIC positioning with repository-layer rebalancing.**

- **Storage:** `columns.position` and `cards.position` become `NUMERIC` (was `INTEGER`). Go type is `float64` across entities and repository signatures.
- **Constants (domain-level):** `STEP = 1000` (seed/end spacing), `RebalanceThreshold = 0.5` (minimum gap before a rebalance).
- **Midpoint:** computed **client-side**. Between neighbors = `(A+B)/2`; at head = `first−STEP`; at tail = `last+STEP`. Head-inserts legitimately produce `0` and negative values — a position is a coordinate on an infinite line, not an index.
- **Move = one UPDATE.** `Move` / `UpdateColumnPosition` write the single trusted `position`. The index-clamp logic (`GetMaxPosition`, `to_position > max+1` clamp) and the neighbor-shift / full-renumber code are deleted.
- **Rebalancing** runs **inline in the move transaction, in the repository.** After the write, inspect only the moved row's two immediate neighbor gaps (two `(column_id, position)` index lookups); if `min(gap) < RebalanceThreshold`, rewrite that column's positions to evenly-spaced `STEP` multiples in the same transaction. A shared internal helper serves both card and column repositories.
- **Validation:** `to_position` / `Position` become `*float64`, **presence-required** — nil → reject, any finite value accepted. `min=0` and `required` are both wrong (they reject the legal value `0` and legal negatives). NaN/±Inf are already rejected by the JSON decoder.
- **Deterministic ordering:** fetch queries use `ORDER BY position ASC, id ASC`. The `id` tiebreaker makes a transient position tie resolve identically on every fetch, so briefly-colliding rows never swap order between reads.
- **No broadcasting in this step.** WebSocket is the next build step; these handlers stay REST-only for now and will be wrapped by the realtime layer (SRS §5.1).

### Migration

- **`up`:** `ALTER … TYPE NUMERIC`, then **rewrite existing rows** to spaced values via `ROW_NUMBER() OVER (PARTITION BY column_id/board_id ORDER BY position, id) * STEP`. The rewrite makes the migration correct against *any* database state (not just an empty one) and prevents split-brain between old gap-1 rows and new gap-1000 seeds — a real footgun during testing, even though there is no production data to preserve.
- **`down`:** `ALTER … TYPE INTEGER USING position::integer` **and re-seed** to contiguous `0,1,2,…` via the same `ROW_NUMBER()` pattern, in one migration. The type reversal alone is lossy (fractions truncate, and truncation can collide); bundling the re-seed makes `down` deterministic and self-healing rather than "lossy, remember to clean up."

## Consequences

**Positive**

- One UPDATE per move; no neighbor is ever touched. No whole-column lock — concurrency-friendly.
- Column reorder can no longer lose updates: two concurrent reorders write different single rows.
- Broadcast payload is a single `{ id, position }` — exactly the WebSocket contract, no per-move event fan-out.
- Positioning internals (`float64`, `STEP`, rebalancing) stay in the repository; the use case and future strategy swaps are insulated.

**Negative / notes**

- **Negative and zero positions are valid** and expected from repeated head-inserts. Any validation or test that assumes positions are non-negative is wrong.
- **Rebalancing rewrites a whole column.** Rare under normal use, but when the WebSocket layer arrives it must handle a bulk "many cards changed" event, not only single moves — the realtime design must account for this.
- **`float64` is a deliberate bound, not arbitrary precision.** The choice is only safe *because* rebalancing caps bisection depth well below `float64`'s limit. If `RebalanceThreshold` were ever lowered toward `float64`'s precision floor, revisit the type.
- **`down` loses exact fractional values** (order is preserved via the re-seed). Acceptable: no production data, and a rollback would re-seed regardless.
