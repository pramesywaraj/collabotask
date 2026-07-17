package card_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/card"
	"collabotask/internal/usecase/common"
)

func TestDeleteCard(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()
	cardID := uuid.New()
	requesterID := uuid.New()

	board := &entity.Board{ID: boardID}
	existingCard := &entity.Card{ID: cardID, ColumnID: columnID}
	column := &entity.Column{ID: columnID, BoardID: boardID}

	validInput := card.DeleteCardInput{
		BoardID:     boardID,
		ColumnID:    columnID,
		CardID:      cardID,
		RequesterID: requesterID,
	}

	tests := []struct {
		name       string
		input      card.DeleteCardInput
		setupMocks func(d cardTestDeps)
		wantErr    error
		wantErrMsg string
	}{
		{
			name:       "invalid input → validation error",
			input:      card.DeleteCardInput{BoardID: boardID, ColumnID: columnID, CardID: cardID}, // missing RequesterID
			setupMocks: func(d cardTestDeps) {},
			wantErrMsg: "validation",
		},
		{
			name:  "card not found → ErrCardNotFound",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(nil, domain.ErrCardNotFound)
			},
			wantErr: domain.ErrCardNotFound,
		},
		{
			name:  "cardRepo.GetByID returns unexpected error → wrapped error",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch card",
		},
		{
			name:  "card belongs to different column → ErrCardNotInColumn",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(
					&entity.Card{ID: cardID, ColumnID: uuid.New()}, nil,
				)
			},
			wantErr: domain.ErrCardNotInColumn,
		},
		{
			name:  "column not found → ErrColumnNotFound",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(nil, domain.ErrColumnNotFound)
			},
			wantErr: domain.ErrColumnNotFound,
		},
		{
			name:  "columnRepo.GetByID returns unexpected error → wrapped error",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch column",
		},
		{
			name:  "column belongs to different board → ErrColumnNotInBoard",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(
					&entity.Column{ID: columnID, BoardID: uuid.New()}, nil,
				)
			},
			wantErr: domain.ErrColumnNotInBoard,
		},
		{
			name:  "boardAccessChecker.Check fails → error propagated",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(column, nil)
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardAccessDenied)
			},
			wantErr: domain.ErrBoardAccessDenied,
		},
		{
			name:  "cardRepo.Delete fails → error",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(column, nil)
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: board}, nil)
				d.cardRepo.EXPECT().Delete(mock.Anything, cardID).Return(errors.New("db error"))
			},
			wantErrMsg: "db error",
		},
		{
			name:  "success",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(column, nil)
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: board}, nil)
				d.cardRepo.EXPECT().Delete(mock.Anything, cardID).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newDeps(t)
			d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
			tt.setupMocks(d)

			err := d.uc.DeleteCard(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			if tt.wantErrMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErrMsg)
				return
			}

			require.NoError(t, err)
		})
	}
}
