package domain

import "errors"

var (
	// Auth
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")

	// Workspace
	ErrMemberNotFound              = errors.New("member not found")
	ErrUserNotInWorkspace          = errors.New("user not in workspace")
	ErrAlreadyMember               = errors.New("user already in workspace")
	ErrNotWorkspaceAdmin           = errors.New("requester is not workspace admin")
	ErrNotWorkspaceOwner           = errors.New("requester is not workspace owner")
	ErrWorkspaceNotFound           = errors.New("workspace not found")
	ErrCannotRemoveYourself        = errors.New("cannot remove yourself")
	ErrBoardOwnerCannotLeave       = errors.New("board owner cannot leave without transferring ownership")
	ErrWorkspaceOwnerCannotLeave   = errors.New("workspace owner cannot leave; transfer ownership or delete the workspace")
	ErrCannotDemoteOwner           = errors.New("cannot demote workspace owner below admin")
	ErrCannotDemoteLastAdmin       = errors.New("cannot demote the last admin; promote another member first")
	ErrLastAdminCannotLeave        = errors.New("last admin cannot leave; promote another member first")

	// Board
	ErrBoardNotFound          = errors.New("board not found")
	ErrBoardAlreadyMember     = errors.New("user already in board")
	ErrBoardMemberNotFound    = errors.New("board member not found")
	ErrBoardAccessDenied      = errors.New("board access denied")
	ErrBoardPermissionDenied  = errors.New("board permission denied")
	ErrBoardCannotJoin        = errors.New("cannot join board, permission denied")
	ErrBoardJoinRequired      = errors.New("join the board before accessing its content")
	ErrBoardNoMembersToInvite        = errors.New("no members were added to the board")
	ErrTransferTargetNotBoardMember  = errors.New("transfer target must be a board member")

	// Column
	ErrColumnNotFound   = errors.New("column not found")
	ErrColumnNotInBoard = errors.New("column not in the board")

	// Card
	ErrCardNotFound      = errors.New("card not found")
	ErrCardNotInColumn   = errors.New("card not in the column")
	ErrInvalidAssigneeID = errors.New("invalid assignee id")

	// Validation
	ErrConstraintViolation = errors.New("constraint violation")
	ErrAtLeastOneProvided  = errors.New("at least provide one of the fields")

	// System
	ErrInconsistentState = errors.New("inconsistent state")
)
