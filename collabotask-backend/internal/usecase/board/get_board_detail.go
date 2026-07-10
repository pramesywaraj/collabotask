package board

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/pkg/validator"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (bu *BoardUseCase) GetBoardDetail(ctx context.Context, input GetBoardDetailInput) (*GetBoardDetailOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate board detail input: %w", err)
	}

	access, err := bu.boardAccessChecker.CheckMetadataAccess(ctx, input.BoardID, input.RequesterID)
	if err != nil {
		return nil, err
	}

	board := access.Board
	joined := access.IsBoardMember()

	// board_members is the sole source of role/access truth (created_by is a
	// stored trace only): a joined requester carries their row's role; anyone
	// else can only ever join (CAN_JOIN). The only non-joined actor that reaches
	// here on a PRIVATE board is a workspace admin — plain members get 404 from
	// the checker.
	var userRole *entity.BoardRole
	accessStatus := entity.BoardJoined
	if joined {
		role := access.BoardMember.Role
		userRole = &role
	} else {
		accessStatus = entity.BoardCanJoin
	}

	// Thin metadata pre-join: hide the roster (member emails) on a PRIVATE board
	// the requester hasn't joined. This only ever affects the non-joined admin;
	// they must Join (break-glass) to see who's on the board. Joined viewers, and
	// everyone on WORKSPACE boards, get the full roster.
	if board.Visibility == entity.BoardVisibilityPrivate && !joined {
		return &GetBoardDetailOutput{
			Board: BoardDetail{
				Board:        board,
				UserRole:     userRole,
				AccessStatus: accessStatus,
				Members:      []BoardMember{},
			},
		}, nil
	}

	members, err := bu.boardMemberRepo.GetMembersByBoard(ctx, input.BoardID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch board members: %w", err)
	}

	userIDs := make([]uuid.UUID, 0, len(members))
	for _, member := range members {
		userIDs = append(userIDs, member.UserID)
	}

	users, err := bu.userRepo.GetByIds(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch members details: %w", err)
	}

	boardMembers := make([]BoardMember, 0, len(members))
	for _, member := range members {
		user, ok := users[member.UserID]
		if !ok || user == nil {
			return nil, domain.ErrUserNotFound
		}
		boardMembers = append(boardMembers, BoardMember{
			UserID:    member.UserID,
			Email:     user.Email,
			Name:      user.Name,
			AvatarURL: user.AvatarURL,
			Role:      member.Role,
			JoinedAt:  member.JoinedAt,
		})
	}

	return &GetBoardDetailOutput{
		Board: BoardDetail{
			Board:        board,
			UserRole:     userRole,
			AccessStatus: accessStatus,
			Members:      boardMembers,
		},
	}, nil
}
