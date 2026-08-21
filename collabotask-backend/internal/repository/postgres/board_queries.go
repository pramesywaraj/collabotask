package postgres

const (
	createBoardQuery = `
		INSERT INTO boards (workspace_id, title, description, created_by, is_archived, background_color, visibility, created_at, updated_at)
		VALUES ($1, $2, $3, $4, FALSE, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, workspace_id, title, description, created_by, is_archived, background_color, visibility, created_at, updated_at
	`
	updateBoardQuery = `
		UPDATE boards
		SET
			title = COALESCE($1, title),
			description = $2,
			background_color = COALESCE($3, background_color),
			visibility = COALESCE($4, visibility),
			updated_at = $5
		WHERE id = $6
		RETURNING id, workspace_id, title, description, created_by, is_archived, background_color, visibility, created_at, updated_at
	`
	getBoardByIDQuery = `
		SELECT id, workspace_id, title, description, created_by, is_archived, background_color, visibility, created_at, updated_at
		FROM boards
		WHERE id = $1
	`
	// Visibility drives the list: a board is shown if the requester is a
	// workspace admin, the board is WORKSPACE-visible, or they are already a
	// member. created_by is no longer consulted — board_members is the sole
	// source of role/access truth (ADR-005). A non-member on a board they may
	// join (admin on any, member on WORKSPACE) gets access_status CAN_JOIN.
	getUserBoardsInWorkspace = `
		SELECT
			b.id, b.workspace_id, b.title, b.description, b.created_by,
			b.is_archived, b.background_color, b.visibility, b.created_at, b.updated_at,
			CASE
				WHEN bm.user_id IS NOT NULL THEN bm.role
				ELSE NULL
			END AS user_role,
			CASE
				WHEN bm.user_id IS NOT NULL THEN 'JOINED'
				WHEN wm.role = 'ADMIN' THEN 'CAN_JOIN'
				WHEN b.visibility = 'WORKSPACE' THEN 'CAN_JOIN'
				ELSE NULL
			END AS access_status,
			COUNT(DISTINCT bm2.user_id)::bigint AS member_count,
			COUNT(DISTINCT c.id)::bigint AS card_count
		FROM boards b
		INNER JOIN workspace_members wm ON b.workspace_id = wm.workspace_id AND wm.user_id = $2
		LEFT JOIN board_members bm ON b.id = bm.board_id AND bm.user_id = $2
		LEFT JOIN board_members bm2 ON b.id = bm2.board_id
		LEFT JOIN columns col ON col.board_id = b.id
		LEFT JOIN cards c ON c.column_id = col.id
		WHERE b.workspace_id = $1
			AND b.is_archived = FALSE
			AND (
				wm.role = 'ADMIN'
				OR b.visibility = 'WORKSPACE'
				OR bm.user_id IS NOT NULL
			)
		GROUP BY b.id, b.workspace_id, b.title, b.description, b.created_by, b.is_archived, b.background_color, b.visibility, b.created_at, b.updated_at, wm.role, bm.role, bm.user_id
		ORDER BY b.created_at DESC
	`
	setBoardArchivedQuery = `
		UPDATE boards SET is_archived = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, workspace_id, title, description, created_by, is_archived, background_color, visibility, created_at, updated_at
	`
	// getBoardIDsByWorkspaceQuery lists every board id in a workspace (archived
	// included). Used by the participation-cascade fan-out to evict a removed user
	// from all workspace board rooms — including WORKSPACE-visible boards they
	// watched without joining (not captured by AffectedBoardIDs). Evicting a room
	// the user is not in is a harmless no-op.
	getBoardIDsByWorkspaceQuery = `
		SELECT id FROM boards WHERE workspace_id = $1
	`
)
