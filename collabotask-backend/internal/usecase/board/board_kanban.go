package board

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/dto"
	"collabotask/internal/infrastructure/validator"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

func (bu *BoardUseCaseImpl) BoardKanban(ctx context.Context, input BoardKanbanInput) (*BoardKanbanOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate board kanban input: %w", err)
	}

	board, err := bu.boardRepo.GetByID(ctx, input.BoardID)
	if err != nil || board == nil {
		if errors.Is(err, domain.ErrBoardNotFound) {
			return nil, domain.ErrBoardNotFound
		}
		return nil, fmt.Errorf("failed to fetch board: %w", err)
	}
	if board.IsArchived {
		return nil, domain.ErrBoardNotFound
	}

	workspaceMembership, err := bu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, board.WorkspaceID, input.RequesterID)
	if err != nil || workspaceMembership == nil || workspaceMembership.IsEmpty() {
		return nil, domain.ErrUserNotInWorkspace
	}

	boardMembership, err := bu.boardMemberRepo.GetMemberByBoardAndUser(ctx, input.BoardID, input.RequesterID)
	if err != nil {
		if !errors.Is(err, domain.ErrBoardMemberNotFound) {
			return nil, fmt.Errorf("failed to fetch board membership: %w", err)
		}

		boardMembership = nil
	}

	hasAccess := workspaceMembership.IsAdmin() || board.CreatedBy == input.RequesterID || (boardMembership != nil && !boardMembership.IsEmpty())
	if !hasAccess {
		return nil, domain.ErrBoardAccessDenied
	}

	columns, err := bu.columnRepo.ListByBoard(ctx, input.BoardID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch list of columns in the board: %w", err)
	}

	var assigneeIDs []uuid.UUID
	// Used to gather which user ids that has been
	// collected, remove any duplicate id that will
	// fetch to the repository.
	// The use of struct{} means that the seen for certain id
	// has been fulfilled and carry no data in it (has 0 size in memory)
	seen := make(map[uuid.UUID]struct{})
	cardsByColumn := make([][]*entity.Card, len(columns))

	for i, col := range columns {
		cards, err := bu.cardRepo.ListByColumn(ctx, col.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to list cards: %w", err)
		}

		cardsByColumn[i] = cards
		for _, card := range cards {
			if card.AssignedTo == nil {
				continue
			}

			id := *card.AssignedTo
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			assigneeIDs = append(assigneeIDs, id)
		}
	}

	users, err := bu.userRepo.GetByIds(ctx, assigneeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch assignees: %w", err)
	}

	out := make([]dto.ColumnWithCardsDTO, len(columns))
	for i, col := range columns {
		dtos := make([]dto.CardWithAssigneeDTO, 0, len(cardsByColumn[i]))
		for _, card := range cardsByColumn[i] {
			var u *entity.User
			if card.AssignedTo != nil {
				if user, ok := users[*card.AssignedTo]; ok {
					u = user
				}
			}
			dtos = append(dtos, dto.CardWithAssigneeToDTO(card, u))
		}
		out[i] = dto.ColumnWithCardsDTO{
			ColumnDTO: dto.ColumnToDTO(col),
			Cards:     dtos,
		}
	}

	return &BoardKanbanOutput{Columns: out}, nil
}
