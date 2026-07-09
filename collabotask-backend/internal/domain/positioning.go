package domain

// Fractional positioning constants (ADR-004). Single source of truth shared by
// the use-case append path (seed spacing) and the repository rebalance path
// (end spacing + rebalance trigger). These MUST stay equal across both layers or
// ordering drifts silently, so they live in the domain layer rather than being
// duplicated per package. The migration seeds with the same STEP value as a
// literal (SQL cannot import Go consts) — keep it in sync.
const (
	// PositionStep is the gap between seeded/appended positions and the spacing
	// used when a partition is rebalanced.
	PositionStep = float64(1000)
	// PositionRebalanceThreshold is the minimum neighbor gap tolerated before a
	// move triggers a full-partition rebalance.
	PositionRebalanceThreshold = float64(0.5)
)
