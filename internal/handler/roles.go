package handler

import (
	"context"
	"errors"
	"time"

	authv1 "auth-service/gen/auth/v1"
	"auth-service/internal/model"

	grpccrud "github.com/Mognus/go-grpc-crud/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

var roleListConfig = grpccrud.ListConfig{
	Searchable:      []string{"name"},
	SortableColumns: []string{"id", "name", "created_at", "updated_at"},
	DefaultSort:     "id ASC",
}

func (h *Handler) GetRole(ctx context.Context, req *authv1.GetRoleRequest) (*authv1.GetRoleResponse, error) {
	role, err := grpccrud.DefaultGet[model.Role](ctx, h.db, req.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "role not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.GetRoleResponse{Role: toRoleResponse(&role)}, nil
}

func (h *Handler) ListRoles(ctx context.Context, req *authv1.ListRolesRequest) (*authv1.ListRolesResponse, error) {
	roles, total, err := grpccrud.DefaultList[model.Role](ctx, h.db, grpccrud.ListRequest{
		Page: req.Page, Limit: req.Limit, Search: req.Search,
		Filters: req.Filters, SortBy: req.SortBy, SortOrder: req.SortOrder,
	}, roleListConfig)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := &authv1.ListRolesResponse{Total: total, Roles: make([]*authv1.RoleResponse, len(roles))}
	for i, r := range roles {
		resp.Roles[i] = toRoleResponse(&r)
	}
	return resp, nil
}

func (h *Handler) CreateRole(ctx context.Context, req *authv1.CreateRoleRequest) (*authv1.CreateRoleResponse, error) {
	role, err := grpccrud.DefaultCreate(ctx, h.db, &model.Role{Name: req.Name})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.CreateRoleResponse{Role: toRoleResponse(role)}, nil
}

func (h *Handler) UpdateRole(ctx context.Context, req *authv1.UpdateRoleRequest) (*authv1.UpdateRoleResponse, error) {
	role, err := grpccrud.DefaultUpdate[model.Role](ctx, h.db, req.Id, map[string]any{"name": req.Name})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "role not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.UpdateRoleResponse{Role: toRoleResponse(role)}, nil
}

func (h *Handler) DeleteRole(ctx context.Context, req *authv1.DeleteRoleRequest) (*authv1.DeleteRoleResponse, error) {
	if err := grpccrud.DefaultDelete(ctx, h.db, &model.Role{}, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.DeleteRoleResponse{Success: true}, nil
}

func toRoleResponse(r *model.Role) *authv1.RoleResponse {
	if r == nil {
		return &authv1.RoleResponse{}
	}
	return &authv1.RoleResponse{
		Id:        uint64(r.ID),
		Name:      r.Name,
		CreatedAt: r.CreatedAt.Format(time.RFC3339),
		UpdatedAt: r.UpdatedAt.Format(time.RFC3339),
	}
}
