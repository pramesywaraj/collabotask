package interfaces

import (
	"collabotask/internal/domain/entities"
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error

	GetById(ctx context.Context, id uuid.UUID) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)

	Update(ctx context.Context, user *entities.User) error

	Delete(ctx context.Context, id uuid.UUID) error

	List(ctx context.Context, limit, offset int) ([]*entities.User, error)

	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
