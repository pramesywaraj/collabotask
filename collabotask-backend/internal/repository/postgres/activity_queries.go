package postgres

const (
	insertActivityQuery = `
		INSERT INTO activities (board_id, user_id, action_type, entity_type, entity_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
)
