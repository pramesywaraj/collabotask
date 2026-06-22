package auth

import (
	"collabotask/internal/config"
	"collabotask/internal/usecase/auth"

	"github.com/google/uuid"
)

var (
	_ auth.PasswordHasher = (*BcryptHasher)(nil)
	_ auth.TokenGenerator = (*JWTGenerator)(nil)
)

type BcryptHasher struct {
	cfg *config.AuthConfig
}

func NewBcryptHasher(cfg *config.AuthConfig) *BcryptHasher {
	return &BcryptHasher{cfg: cfg}
}

func (h *BcryptHasher) Hash(password string) (string, error) {
	return HashPassword(h.cfg, password)
}

func (h *BcryptHasher) Check(password, hash string) bool {
	return CheckPassword(password, hash)
}

type JWTGenerator struct {
	cfg *config.AuthConfig
}

func NewJWTGenerator(cfg *config.AuthConfig) *JWTGenerator {
	return &JWTGenerator{cfg: cfg}
}

func (g *JWTGenerator) Generate(userID uuid.UUID, role string) (string, error) {
	return GenerateToken(g.cfg, userID, role)
}
