package postgres

const (
	createWorkspaceMemberQuery = `
		INSERT INTO workspace_members (workspace_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		RETURNING workspace_id, user_id, role, joined_at
	`
	deleteWorkspaceMemberQuery = `
		DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2
	`
	getByWorkspaceAndUserQuery = `
		SELECT workspace_id, user_id, role, joined_at FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2
	`
	listMemberByWorkspaceQuery = `
		SELECT workspace_id, user_id, role, joined_at FROM workspace_members
		WHERE workspace_id = $1 ORDER BY joined_at ASC
	`
	isUserExistsOnWorkspaceQuery = `
		SELECT EXISTS(
			SELECT 1
			FROM workspace_members
			WHERE workspace_id = $1 AND user_id = $2
		)
	`
)
