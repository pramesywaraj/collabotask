package postgres

const (
	createWorkspaceQuery = `
		INSERT INTO workspaces (name, description, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, name, description, owner_id, created_at, updated_at
	`
	updateWorkspaceQuery = `
		UPDATE workspaces
		SET
			name = COALESCE($1, name),
			description = COALESCE($2, description),
			updated_at = $3
		WHERE id = $4
		RETURNING id, name, description, owner_id, created_at, updated_at
	`
	deleteWorkspaceQuery = `
		DELETE FROM workspaces WHERE id = $1
	`
	getWorkspaceByIdQuery = `
		SELECT id, name, description, owner_id, created_at, updated_at FROM workspaces
		WHERE id = $1
	`
	getUserWorkspacesQuery = `
		SELECT
			w.id, w.name, w.description, w.owner_id, w.created_at, w.updated_at,
			wm.role AS role,
			COUNT(DISTINCT wm2.user_id) AS member_count,
			0::bigint AS board_count
		FROM workspaces w
		INNER JOIN workspace_members wm ON w.id = wm.workspace_id AND wm.user_id = $1
		LEFT JOIN workspace_members wm2 ON w.id = wm2.workspace_id
		WHERE wm.user_id = $1
		GROUP BY w.id, w.name, w.description, w.owner_id, w.created_at, w.updated_at, wm.role
		ORDER BY w.created_at DESC
	`
)
