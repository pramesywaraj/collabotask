package postgres

const (
	createBoardMemberQuery = `
		INSERT INTO board_members (board_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
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
