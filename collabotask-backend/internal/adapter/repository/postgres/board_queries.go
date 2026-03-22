package postgres

const (
	createBoardQuery = `
		INSERT INTO boards (workspace_id, title, description, created_by, is_archived, background_color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, FALSE, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, workspace_id, title, description, created_by, is_archived, background_color, created_at, updated_at
	`
	updateBoardQuery = `
		UPDATE boards
		SET
			title = COALESCE($1, title),
			description = COALESCE($2, description),
			background_color = COALESCE($3, background_color),
			updated_at = $4
		WHERE id = $5
		RETURNING id, workspace_id, title, description, created_by, is_archived, background_color, created_at, updated_at
	`
	deleteBoardQuery = `
		DELETE FROM boards WHERE id = $1
	`
	getBoardByIDQuery = `
		SELECT id, workspace_id, title, description, created_by, is_archived, background_color, created_at, updated_at
		FROM boards
		WHERE id = $1
	`
	getUserBoardsInWorkspace = `
		SELECT
			b.id, b.workspace_id, b.title, b.description, b.created_by,
			b.is_archived, b.background_color, b.created_at, b.updated_at,
			CASE
				WHEN bm.user_id IS NOT NULL THEN bm.role
				WHEN b.created_by = $2 THEN 'BOARD_OWNER'
				ELSE NULL
			END AS user_role,
			CASE
				WHEN bm.user_id IS NOT NULL OR b.created_by = $2 THEN 'JOINED'
				WHEN wm.role = 'ADMIN' THEN 'CAN_JOIN'
				ELSE NULL
			END AS access_status,
			COUNT(DISTINCT bm2.user_id)::bigint AS member_count
		FROM boards b
		INNER JOIN workspace_members wm ON b.workspace_id = wm.workspace_id AND wm.user_id = $2
		LEFT JOIN board_members bm ON b.id = bm.board_id AND bm.user_id = $2
		LEFT JOIN board_members bm2 ON b.id = bm2.board_id
		WHERE b.workspace_id = $1
			AND b.is_archived = FALSE
			AND (
				wm.role = 'ADMIN'
				OR b.created_by = $2
				OR bm.user_id IS NOT NULL
			)
		GROUP BY b.id, b.workspace_id, b.title, b.description, b.created_by, b.is_archived, b.background_color, b.created_at, b.updated_at, wm.role, bm.role, bm.user_id
		ORDER BY b.created_at DESC
	`
	setBoardArchivedQuery = `
		UPDATE boards SET is_archived = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, workspace_id, title, description, created_by, is_archived, background_color, created_at, updated_at
	`
)
