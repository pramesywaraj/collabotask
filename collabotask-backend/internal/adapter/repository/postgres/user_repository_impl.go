package postgres

import (
	"collabotask/internal/domain/entity"
	"collabotask/internal/domain/repository"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	var avatarUrl *string
	if user.AvatarURL != nil && *user.AvatarURL != "" {
		avatarUrl = user.AvatarURL
	}

	err := r.db.QueryRow(
		ctx,
		createUserQuery,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.SystemRole,
		avatarUrl,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.AvatarURL,
		&user.SystemRole,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return fmt.Errorf("email already exists")
			}
		}

		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepositoryImpl) GetById(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user := &entity.User{}
	var avatarUrl *string

	err := r.db.QueryRow(ctx, getUserByIdQuery, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&avatarUrl,
		&user.SystemRole,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	user.AvatarURL = avatarUrl
	return user, nil
}

func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	user := &entity.User{}
	var avatarUrl *string

	err := r.db.QueryRow(ctx, getUserByEmailQuery, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&avatarUrl,
		&user.SystemRole,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	user.AvatarURL = avatarUrl
	return user, nil
}

func (r *UserRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
	var avatarUrl *string
	if user.AvatarURL != nil && *user.AvatarURL != "" {
		avatarUrl = user.AvatarURL
	}

	var email *string
	if user.Email != "" {
		email = &user.Name
	}

	var name *string
	if user.Name != "" {
		name = &user.Name
	}

	var passwordHash *string
	if user.PasswordHash != "" {
		passwordHash = &user.PasswordHash
	}

	err := r.db.QueryRow(
		ctx,
		updateUserQuery,
		email,
		name,
		avatarUrl,
		passwordHash,
		user.UpdatedAt,
		user.ID,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.AvatarURL,
		&user.SystemRole,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("user not found")
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return fmt.Errorf("email already exists")
			}
		}

		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *UserRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx, deleteUserQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *UserRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	rows, err := r.db.Query(ctx, listUsersQuery, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := []*entity.User{}
	for rows.Next() {
		user := &entity.User{}
		var avatarUrl *string

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&avatarUrl,
			&user.SystemRole,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		user.AvatarURL = avatarUrl
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (r *UserRepositoryImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, existsUserByEmailQuery, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return exists, nil
}
