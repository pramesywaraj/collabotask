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
	updateWorkspaceMemberRoleQuery = `
		UPDATE workspace_members SET role = $3
		WHERE workspace_id = $1 AND user_id = $2
		RETURNING workspace_id, user_id, role, joined_at
	`
	countWorkspaceAdminsQuery = `
		SELECT COUNT(*) FROM workspace_members
		WHERE workspace_id = $1 AND role = 'ADMIN'
	`
	// Participation cascade (step 1 reuses deleteWorkspaceMemberQuery above).
	// RETURNING bm.board_id captures every board the user was a member of so the
	// caller can write one MEMBER/REMOVED activity row per board (ADR-007, §4.6).
	deleteBoardMembershipsForUserQuery = `
		DELETE FROM board_members bm
		USING boards b
		WHERE bm.board_id = b.id AND b.workspace_id = $1 AND bm.user_id = $2
		RETURNING bm.board_id
	`
	unassignCardsForUserQuery = `
		UPDATE cards c SET assigned_to = NULL
		FROM columns col
		JOIN boards b ON col.board_id = b.id
		WHERE c.column_id = col.id AND b.workspace_id = $1 AND c.assigned_to = $2
		RETURNING c.id, c.column_id, b.id
	`
)
