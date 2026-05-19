package service

import (
	"context"
	"errors"

	"auth-service/internal/domain/users"

	"github.com/Mognus/go-grpc-crud/dbcrud"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var userListConfig = dbcrud.ListConfig{
	Preloads:        []string{"Role"},
	Searchable:      []string{"email", "first_name", "last_name"},
	Filterable:      []string{"id", "email", "firstName", "lastName", "active", "roleId"},
	SortableColumns: []string{"id", "email", "firstName", "lastName", "active", "createdAt", "updatedAt"},
	DefaultSort:     "id ASC",
}

type UserService struct {
	db *gorm.DB
}

type CreateUserInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	RoleID    uint
	Active    bool
}

type UpdateUserInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	RoleID    uint64
	Active    bool
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) Get(ctx context.Context, id uint64) (users.User, error) {
	return dbcrud.DefaultGet[users.User](ctx, s.db, id, "Role")
}

func (s *UserService) List(ctx context.Context, req dbcrud.ListRequest) ([]users.User, int64, error) {
	return dbcrud.DefaultList[users.User](ctx, s.db, req, userListConfig)
}

func (s *UserService) Create(ctx context.Context, input CreateUserInput) (*users.User, error) {
	user, err := dbcrud.DefaultCreate(ctx, s.db, &users.User{
		Email:     input.Email,
		Password:  input.Password,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		RoleID:    input.RoleID,
		Active:    &input.Active,
	}, "Role")
	if err != nil {
		if isUniqueConstraintViolation(err, "users_email_key") {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	return user, nil
}

func (s *UserService) Update(ctx context.Context, id uint64, input UpdateUserInput) (*users.User, error) {
	updates := map[string]any{
		"email":      input.Email,
		"first_name": input.FirstName,
		"last_name":  input.LastName,
		"role_id":    input.RoleID,
		"active":     input.Active,
	}
	if input.Password != "" {
		updates["password"] = input.Password
	}

	return dbcrud.DefaultUpdate[users.User](ctx, s.db, id, updates, "Role")
}

func (s *UserService) Delete(ctx context.Context, id uint64) error {
	return dbcrud.DefaultDelete(ctx, s.db, &users.User{}, id)
}

func isUniqueConstraintViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == "23505" && pgErr.ConstraintName == constraintName
}
