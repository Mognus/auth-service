package service

import (
	"context"

	"auth-service/internal/domain/users"

	grpccrud "github.com/Mognus/go-grpc-crud/server"
	"gorm.io/gorm"
)

var userListConfig = grpccrud.ListConfig{
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
	return grpccrud.DefaultGet[users.User](ctx, s.db, id, "Role")
}

func (s *UserService) List(ctx context.Context, req grpccrud.ListRequest) ([]users.User, int64, error) {
	return grpccrud.DefaultList[users.User](ctx, s.db, req, userListConfig)
}

func (s *UserService) Create(ctx context.Context, input CreateUserInput) (*users.User, error) {
	return grpccrud.DefaultCreate(ctx, s.db, &users.User{
		Email:     input.Email,
		Password:  input.Password,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		RoleID:    input.RoleID,
		Active:    input.Active,
	}, "Role")
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

	return grpccrud.DefaultUpdate[users.User](ctx, s.db, id, updates, "Role")
}

func (s *UserService) Delete(ctx context.Context, id uint64) error {
	return grpccrud.DefaultDelete(ctx, s.db, &users.User{}, id)
}
