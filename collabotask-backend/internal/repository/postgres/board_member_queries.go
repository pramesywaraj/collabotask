package postgres

const (
	createBoardMemberQuery = `
		INSERT INTO board_members (board_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
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
)
