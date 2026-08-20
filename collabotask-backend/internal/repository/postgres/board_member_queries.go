package postgres

const (
	createBoardMemberQuery = `
		INSERT INTO board_members (board_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		RETURNING joined_at
	`
	// createBoardMemberIfAbsentQuery is idempotent: a returned row means the
	// member was newly inserted; no row (ErrNoRows) means they were already a
	// member. This drives self-join's "joined" flag (ADR-005).
	createBoardMemberIfAbsentQuery = `
		INSERT INTO board_members (board_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (board_id, user_id) DO NOTHING
		RETURNING board_id
	`
	deleteBoardMemberQuery = `
		DELETE FROM board_members WHERE board_id = $1 AND user_id = $2
	`
	listMemberByBoardQuery = `
		SELECT
			board_id, user_id, role, joined_at
		FROM board_members
		WHERE board_id = $1
		ORDER BY joined_at ASC
	`
	isUserExistsOnBoardQuery = `
		SELECT EXISTS(
			SELECT 1
			FROM board_members
			WHERE board_id = $1 AND user_id = $2
		)
	`
	getMemberByBoardAndUserQuery = `
		SELECT
			board_id, user_id, role, joined_at
		FROM board_members
		WHERE board_id = $1 AND user_id = $2
	`
	// demoteCurrentOwnerQuery demotes the current owner by role and returns their
	// user_id. Returns 0 rows (pgx.ErrNoRows) when the board has no owner (orphan).
	demoteCurrentOwnerQuery = `
		UPDATE board_members SET role = 'BOARD_MEMBER'
		WHERE board_id = $1 AND role = 'BOARD_OWNER'
		RETURNING user_id
	`
	// promoteNewOwnerQuery promotes the target member to BOARD_OWNER by user id.
	promoteNewOwnerQuery = `
		UPDATE board_members SET role = 'BOARD_OWNER'
		WHERE board_id = $1 AND user_id = $2
	`
	// deleteBoardMemberForCascadeQuery removes a board_members row; 0 rows → ErrBoardMemberNotFound.
	deleteBoardMemberForCascadeQuery = `
		DELETE FROM board_members WHERE board_id = $1 AND user_id = $2
	`
	// unassignBoardCardsForUserQuery clears card assignments for a user leaving a board.
	// Uses columns.board_id so no board join is needed on the cards table.
	unassignBoardCardsForUserQuery = `
		UPDATE cards c SET assigned_to = NULL
		FROM columns col
		WHERE c.column_id = col.id AND col.board_id = $1 AND c.assigned_to = $2
		RETURNING c.id, c.column_id
	`
)
