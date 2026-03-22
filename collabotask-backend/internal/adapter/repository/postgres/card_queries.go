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
		ORDER BY position ASC
	`
	getMaxCardPositionQuery = `
		SELECT COALESCE(MAX(position), -1)
		FROM cards
		WHERE column_id = $1
	`
	updateCardQuery = `
		UPDATE cards
		SET
			title = COALESCE($1, title),
			description = COALESCE($2, description),
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
	incrementPositionCardFromQuery = `
		UPDATE cards
		SET
			position = position + 1
		WHERE column_id = $1 AND position >= $2
	`
	decrementPositionCardAfterQuery = `
		UPDATE cards
		SET
			position = position - 1
		WHERE column_id = $1 AND position > $2
	`
)
