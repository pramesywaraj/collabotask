package domain

import "errors"

var (
	// Auth
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")

	// Workspace
	ErrMemberNotFound       = errors.New("member not found")
	ErrUserNotInWorkspace   = errors.New("user not in workspace")
	ErrAlreadyMember        = errors.New("user already in workspace")
	ErrNotWorkspaceAdmin    = errors.New("requester is not workspace admin")
	ErrWorkspaceNotFound    = errors.New("workspace not found")
	ErrCannotRemoveYourself = errors.New("cannot remove yourself")

	// Board
	ErrBoardNotFound         = errors.New("board not found")
	ErrBoardAlreadyMember    = errors.New("user already in board")
	ErrBoardMemberNotFound   = errors.New("board member not found")
	ErrBoardAccessDenied     = errors.New("board access denied")
	ErrBoardPermissionDenied = errors.New("board permission denied")

	// Validation
	ErrConstraintViolation = errors.New("constraint violation")
)
