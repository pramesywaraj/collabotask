package board

import (
	"collabotask/internal/domain/entity"
	"collabotask/pkg/validator"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (bu *BoardUseCase) GetBoardKanban(ctx context.Context, input GetBoardKanbanInput) (*GetBoardKanbanOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate board kanban input: %w", err)
	}

	if _, err := bu.boardAccessChecker.CheckViewAccess(ctx, input.BoardID, input.RequesterID); err != nil {
		return nil, err
	}

	columns, err := bu.columnRepo.GetColumnsByBoard(ctx, input.BoardID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch list of columns in the board: %w", err)
	}

	var assigneeIDs []uuid.UUID
	seen := make(map[uuid.UUID]struct{})
	cardsByColumn := make([][]*entity.Card, len(columns))

	for i, col := range columns {
		cards, err := bu.cardRepo.GetCardsByColumn(ctx, col.ID)
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

	out := make([]ColumnWithCards, len(columns))
	for i, col := range columns {
		cards := make([]CardWithAssignee, 0, len(cardsByColumn[i]))
		for _, card := range cardsByColumn[i] {
			var u *entity.User
			if card.AssignedTo != nil {
				if user, ok := users[*card.AssignedTo]; ok {
					u = user
				}
			}
			cards = append(cards, CardWithAssignee{Card: card, Assignee: u})
		}
		out[i] = ColumnWithCards{
			Column: col,
			Cards:  cards,
		}
	}

	return &GetBoardKanbanOutput{Columns: out}, nil
}
