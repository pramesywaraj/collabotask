# Testing Playbook — collabotask-backend

Tooling decision: **testify + mockery**. See [ADR-003](../docs/architecture/adr/adr-003-unit-testing-strategy.md) for why.

---

## Where Tests Live (The Pyramid)

```
internal/usecase/     ← write the most tests here (business logic)
internal/domain/      ← pure functions / entity methods (no mocks needed)
internal/delivery/    ← light handler tests (HTTP binding only)
internal/repository/  ← DO NOT unit test; use integration tests with a real DB
```

The use case layer is the highest-value target. It holds all permission checks, branching logic, and error handling. Repository interfaces give us clean seams to inject mocks — no Postgres needed.

---

## Setup

### 1. Install mockery

```bash
brew install mockery          # macOS
# or: go install github.com/vektra/mockery/v2@latest
```

### 2. Install testify

```bash
go get github.com/stretchr/testify
```

### 3. Add `.mockery.yaml` at the backend root

> Requires **mockery v3**. Key names changed from v2: `outpkg` → `pkgname`, `mockname` → `structname`, `with-expecter` is replaced by `template: testify` (expecter is included by default). Each source package needs its own output `filename` to avoid a "same source package" constraint.

```yaml
dir: internal/mocks
structname: Mock{{.InterfaceName}}
pkgname: mocks
template: testify
packages:
  collabotask/internal/domain/repository:
    config:
      filename: repository_mocks.go
    interfaces:
      BoardRepository: {}
      BoardMemberRepository: {}
      WorkspaceMemberRepository: {}
      CardRepository: {}
      ColumnRepository: {}
      UserRepository: {}
  collabotask/internal/usecase/common:
    config:
      filename: common_mocks.go
    interfaces:
      BoardAccessChecker: {}
  collabotask/internal/usecase/auth:
    config:
      filename: auth_mocks.go
    interfaces:
      PasswordHasher: {}
      TokenGenerator: {}
```

### 4. Generate all mocks

```bash
# Run from collabotask-backend/
mockery
```

This creates `internal/mocks/` with one file per interface. **Never edit generated files** — re-run `mockery` instead.

---

## Conventions

| Rule | Detail |
|---|---|
| File name | `update_board_test.go` mirrors `update_board.go` |
| Package | `package board_test` — external test package; tests the public API |
| Mock import | `"collabotask/internal/mocks"` |
| Test function | `TestUpdateBoard`, `TestCreateCard`, etc. |
| Sub-tests | Table-driven with `t.Run(tt.name, ...)` |

---

## Structure of a Use Case Test

Every use case test follows this skeleton:

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name       string
        input      SomeInput
        setupMocks func(dep1 *mocks.MockDep1, dep2 *mocks.MockDep2)
        wantErr    error
        // add wantOutput fields only when asserting the return value
    }{
        { /* case 1 */ },
        { /* case 2 */ },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 1. Create mocks
            dep1 := mocks.NewMockDep1(t)
            dep2 := mocks.NewMockDep2(t)

            // 2. Wire expectations
            tt.setupMocks(dep1, dep2)

            // 3. Build use case and call it
            uc := NewSomeUseCase(dep1, dep2)
            _, err := uc.SomeMethod(context.Background(), tt.input)

            // 4. Assert
            if tt.wantErr != nil {
                require.ErrorIs(t, err, tt.wantErr)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

**`require` vs `assert`**
- `require` stops the test immediately on failure — use it for preconditions (error checks, nil guards)
- `assert` continues after failure — use it for multiple independent field checks on the output

When using `mocks.NewMock*(t)` (passing `t`), mockery registers `AssertExpectations` automatically — you do not need to call it manually.

---

## Worked Example: `UpdateBoard`

File: `internal/usecase/board/update_board_test.go`

```go
package board_test

import (
    "context"
    "errors"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/require"

    "collabotask/internal/domain"
    "collabotask/internal/domain/entity"
    "collabotask/internal/mocks"
    "collabotask/internal/usecase/board"
)

func TestUpdateBoard(t *testing.T) {
    // Shared fixtures — reused across cases
    requesterID  := uuid.New()
    boardID      := uuid.New()
    workspaceID  := uuid.New()
    someoneElse  := uuid.New()
    newTitle     := "Redesigned Title"

    // A board whose creator is someone else (requester must earn access another way)
    aBoard := func() *entity.Board {
        return &entity.Board{
            ID:          boardID,
            WorkspaceID: workspaceID,
            CreatedBy:   someoneElse,
            Title:       "Old Title",
        }
    }

    wsMember := func(role entity.WorkspaceRole) *entity.WorkspaceMember {
        return &entity.WorkspaceMember{
            WorkspaceID: workspaceID,
            UserID:      requesterID,
            Role:        role,
        }
    }

    boardMember := func(role entity.BoardRole) *entity.BoardMember {
        return &entity.BoardMember{
            BoardID: boardID,
            UserID:  requesterID,
            Role:    role,
        }
    }

    tests := []struct {
        name       string
        input      board.UpdateBoardInput
        setupMocks func(
            boardRepo      *mocks.MockBoardRepository,
            wsMemberRepo   *mocks.MockWorkspaceMemberRepository,
            boardMemberRepo *mocks.MockBoardMemberRepository,
        )
        wantErr error
    }{
        {
            name: "no fields provided → ErrAtLeastOneProvided",
            input: board.UpdateBoardInput{
                RequesterID: requesterID,
                BoardID:     boardID,
                // Title, BackgroundColor, DescriptionPresent all zero
            },
            setupMocks: func(br *mocks.MockBoardRepository, ws *mocks.MockWorkspaceMemberRepository, bm *mocks.MockBoardMemberRepository) {
                // Validation fails before any repo call — nothing to set up
            },
            wantErr: domain.ErrAtLeastOneProvided,
        },
        {
            name: "board not found → ErrBoardNotFound",
            input: board.UpdateBoardInput{
                RequesterID: requesterID,
                BoardID:     boardID,
                Title:       &newTitle,
            },
            setupMocks: func(br *mocks.MockBoardRepository, ws *mocks.MockWorkspaceMemberRepository, bm *mocks.MockBoardMemberRepository) {
                br.EXPECT().GetByID(context.Background(), boardID).Return(nil, domain.ErrBoardNotFound)
                // ws and bm: no expectations — they must NOT be called
            },
            wantErr: domain.ErrBoardNotFound,
        },
        {
            name: "requester not in workspace → ErrUserNotInWorkspace",
            input: board.UpdateBoardInput{
                RequesterID: requesterID,
                BoardID:     boardID,
                Title:       &newTitle,
            },
            setupMocks: func(br *mocks.MockBoardRepository, ws *mocks.MockWorkspaceMemberRepository, bm *mocks.MockBoardMemberRepository) {
                br.EXPECT().GetByID(context.Background(), boardID).Return(aBoard(), nil)
                ws.EXPECT().GetByWorkspaceAndUser(context.Background(), workspaceID, requesterID).Return(nil, errors.New("not found"))
            },
            wantErr: domain.ErrUserNotInWorkspace,
        },
        {
            name: "workspace admin can administer without being board member",
            input: board.UpdateBoardInput{
                RequesterID: requesterID,
                BoardID:     boardID,
                Title:       &newTitle,
            },
            setupMocks: func(br *mocks.MockBoardRepository, ws *mocks.MockWorkspaceMemberRepository, bm *mocks.MockBoardMemberRepository) {
                br.EXPECT().GetByID(context.Background(), boardID).Return(aBoard(), nil)
                ws.EXPECT().GetByWorkspaceAndUser(context.Background(), workspaceID, requesterID).Return(wsMember(entity.WorkspaceRoleAdmin), nil)
                bm.EXPECT().GetMemberByBoardAndUser(context.Background(), boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
                br.EXPECT().Update(context.Background(), aBoard()).Return(nil)
            },
            wantErr: nil,
        },
        {
            name: "regular member who is not board owner → ErrBoardPermissionDenied",
            input: board.UpdateBoardInput{
                RequesterID: requesterID,
                BoardID:     boardID,
                Title:       &newTitle,
            },
            setupMocks: func(br *mocks.MockBoardRepository, ws *mocks.MockWorkspaceMemberRepository, bm *mocks.MockBoardMemberRepository) {
                br.EXPECT().GetByID(context.Background(), boardID).Return(aBoard(), nil)
                ws.EXPECT().GetByWorkspaceAndUser(context.Background(), workspaceID, requesterID).Return(wsMember(entity.WorkspaceRoleMember), nil)
                bm.EXPECT().GetMemberByBoardAndUser(context.Background(), boardID, requesterID).Return(boardMember(entity.BoardRoleMember), nil)
            },
            wantErr: domain.ErrBoardPermissionDenied,
        },
        {
            name: "success — board owner updates title",
            input: board.UpdateBoardInput{
                RequesterID: requesterID,
                BoardID:     boardID,
                Title:       &newTitle,
            },
            setupMocks: func(br *mocks.MockBoardRepository, ws *mocks.MockWorkspaceMemberRepository, bm *mocks.MockBoardMemberRepository) {
                br.EXPECT().GetByID(context.Background(), boardID).Return(aBoard(), nil)
                ws.EXPECT().GetByWorkspaceAndUser(context.Background(), workspaceID, requesterID).Return(wsMember(entity.WorkspaceRoleMember), nil)
                bm.EXPECT().GetMemberByBoardAndUser(context.Background(), boardID, requesterID).Return(boardMember(entity.BoardRoleOwner), nil)
                br.EXPECT().Update(context.Background(), aBoard()).Return(nil)
            },
            wantErr: nil,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            boardRepo      := mocks.NewMockBoardRepository(t)
            wsMemberRepo   := mocks.NewMockWorkspaceMemberRepository(t)
            boardMemberRepo := mocks.NewMockBoardMemberRepository(t)
            userRepo       := mocks.NewMockUserRepository(t)
            columnRepo     := mocks.NewMockColumnRepository(t)
            cardRepo       := mocks.NewMockCardRepository(t)
            accessChecker  := mocks.NewMockBoardAccessChecker(t)

            tt.setupMocks(boardRepo, wsMemberRepo, boardMemberRepo)

            uc := board.NewBoardUseCase(accessChecker, boardRepo, boardMemberRepo, wsMemberRepo, userRepo, columnRepo, cardRepo)
            _, err := uc.UpdateBoard(context.Background(), tt.input)

            if tt.wantErr != nil {
                require.ErrorIs(t, err, tt.wantErr)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

---

## What Each Test Case Should Cover

For each use case, ask: **what are all the ways this can fail?** Write one table row per branch:

- Input validation fails (missing required fields)
- "Not found" for each entity the use case fetches
- Permission check fails
- DB error on the write operation
- Success (happy path — usually last and least interesting)

The error branches are where bugs hide. The happy path is often the least valuable test.

---

## What NOT to Unit Test

| Layer | Why |
|---|---|
| `internal/repository/postgres/` | Talks to Postgres; needs a real DB. Use integration tests |
| `internal/injection/` (Wire) | Generated code; don't test generated code |
| `internal/delivery/http/handler/` | HTTP binding is thin; test it with `httptest` only if JSON mapping is non-trivial |
| `internal/infrastructure/` | Auth (bcrypt, JWT) and DB connection are I/O; test in integration |

---

## Running Tests

```bash
# All tests
go test ./...

# A specific package
go test ./internal/usecase/board/...

# With verbose output
go test -v ./internal/usecase/board/...

# Regenerate mocks after changing an interface
mockery
```

## Check Unit Tests Coverage

```bash
# Quick percentage per package
go test ./internal/usecase/auth/... -cover

# Whole project
go test ./... -cover

# Function-level breakdown
go test ./internal/usecase/auth/... -coverprofile=cov.out
go tool cover -func=cov.out

# Visual HTML report
go test ./internal/usecase/auth/... -coverprofile=cov.out
go tool cover -html=cov.out

# Coverage across the whole module in one profile
go test ./... -coverprofile=cov.out
go tool cover -func=cov.out | tail -1

```
