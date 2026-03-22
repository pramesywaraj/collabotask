package postgres

const (
	createColumnQuery = `
		INSERT INTO columns (board_id, title, position, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, board_id, title, position, created_at, updated_at
	`
	updateColumnQuery = `
		UPDATE columns
		SET
			title = COALESCE($1, title),
			position = COALESCE($2, position),
			updated_at = $3
		WHERE id = $4
		RETURNING id, board_id, title, position, created_at, updated_at
	`
	deleteColumnQuery = `
		DELETE FROM columns WHERE id = $1
	`
	getColumnByIDQuery = `
		SELECT id, board_id, title, position, created_at, updated_at
		FROM columns
		WHERE id = $1
	`
	listColumnByBoardIDQuery = `
		SELECT id, board_id, title, position, created_at, updated_at
		FROM columns
		WHERE board_id = $1
		ORDER BY position ASC
	`
	getColumnMaxPositionQuery = `
		SELECT COALESCE(MAX(position), -1)
		FROM columns
		WHERE board_id = $1
	`
)
