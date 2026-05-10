package service

import (
	"context"

	"auth-service/internal/domain/roles"

	grpccrud "github.com/Mognus/go-grpc-crud/server"
	"gorm.io/gorm"
)

var roleListConfig = grpccrud.ListConfig{
	Searchable:      []string{"name"},
	Filterable:      []string{"name"},
	SortableColumns: []string{"id", "name", "createdAt", "updatedAt"},
	DefaultSort:     "id ASC",
}

type RoleService struct {
	db *gorm.DB
}

func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{db: db}
}

func (s *RoleService) Get(ctx context.Context, id uint64) (roles.Role, error) {
	return grpccrud.DefaultGet[roles.Role](ctx, s.db, id)
}

func (s *RoleService) List(ctx context.Context, req grpccrud.ListRequest) ([]roles.Role, int64, error) {
	return grpccrud.DefaultList[roles.Role](ctx, s.db, req, roleListConfig)
}

func (s *RoleService) Create(ctx context.Context, name string) (*roles.Role, error) {
	return grpccrud.DefaultCreate(ctx, s.db, &roles.Role{Name: name})
}

func (s *RoleService) Update(ctx context.Context, id uint64, name string) (*roles.Role, error) {
	return grpccrud.DefaultUpdate[roles.Role](ctx, s.db, id, map[string]any{"name": name})
}

func (s *RoleService) Delete(ctx context.Context, id uint64) error {
	return grpccrud.DefaultDelete(ctx, s.db, &roles.Role{}, id)
}
