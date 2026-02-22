package postgres

const (
	createWorkspaceQuery = `
		INSERT INTO workspaces (name, description, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3)
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
		SELECT w.id, w.name, w.description, w.owner_id, w.created_at, w.updated_at FROM workspaces w
		INNER JOIN workspace_members wm ON w.id = wm.workspace_id
		WHERE wm.user_id = $1 ORDER BY w.created_at DESC
	`
)
