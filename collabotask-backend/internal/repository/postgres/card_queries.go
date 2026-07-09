package postgres

const (
	createCardQuery = `
		INSERT INTO cards (column_id, title, description, position, assigned_to, due_date, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, column_id, title, description, position, assigned_to, due_date, created_by, created_at, updated_at
	`
	getCardByIDQuery = `
		SELECT id, column_id, title, description, position, assigned_to, due_date, created_by, created_at, updated_at
		FROM cards
		WHERE id = $1
	`
	listCardByColumnQuery = `
		SELECT id, column_id, title, description, position, assigned_to, due_date, created_by, created_at, updated_at
		FROM cards
		WHERE column_id = $1
		ORDER BY position ASC, id ASC
	`
	getMaxCardPositionQuery = `
		SELECT COALESCE(MAX(position), 0)
		FROM cards
		WHERE column_id = $1
	`
	updateCardQuery = `
		UPDATE cards
		SET
			title = COALESCE($1, title),
			description = $2,
			assigned_to = $3,
			due_date = $4,
			updated_at = $5
		WHERE id = $6
		RETURNING id, column_id, title, description, position, assigned_to, due_date, created_by, created_at, updated_at
	`
	deleteCardQuery = `
		DELETE FROM cards
		WHERE id = $1
	`
	lockCardForMoveQuery = `
		SELECT column_id
		FROM cards
		WHERE id = $1
		FOR UPDATE
	`
	moveCardQuery = `
		UPDATE cards
		SET
			column_id = $1,
			position = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING id, column_id, title, description, position, assigned_to, due_date, created_by, created_at, updated_at
	`
)
