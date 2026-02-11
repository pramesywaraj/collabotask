package postgres

const (
	createUserQuery = `
		INSERT INTO users (email, password_hash, name, system_role, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, email, name, avatar_url, system_role, created_at, updated_at
	`
	getUserByIdQuery = `
		SELECT id, email, name, avatar_url, system_role, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	getUserByEmailQuery = `
		SELECT id, email, name, avatar_url, system_role, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	updateUserQuery = `
		UPDATE users
		SET
			email = COALESCE($1, email)
			name = COALESCE($2, name)
			avatar_url = COALESCE($3, avatar_url)
			password_hash = COALESCE($4, password_hash)
			updated_at = $5
		WHERE id = $6
		RETURNING id, email, name, avatar_url, system_role, created_at, updated_at
	`
	deleteUserQuery = `
		DELETE FROM users
		WHERE id = $1
	`
	listUsersQuery = `
		SELECT id, email, name, avatar_url, system_role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	existsUserByEmailQuery = `
		SELECT EXISTS(
			SELECT 1 
			FROM users 
			WHERE email = $1
		)
	`
)
