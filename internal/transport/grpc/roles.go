package grpc

import (
	"context"
	"errors"
	"time"

	authv1 "auth-service/gen/auth/v1"
	"auth-service/internal/domain/roles"

	"github.com/Mognus/go-grpc-crud/dbcrud"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func (h *Handler) GetRole(ctx context.Context, req *authv1.GetRoleRequest) (*authv1.GetRoleResponse, error) {
	role, err := h.roleService.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "role")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.GetRoleResponse{Role: toRoleResponse(&role)}, nil
}

func (h *Handler) ListRoles(ctx context.Context, req *authv1.ListRolesRequest) (*authv1.ListRolesResponse, error) {
	result, total, err := h.roleService.List(ctx, dbcrud.ListRequest{
		Page: req.Page, Limit: req.Limit, Search: req.Search,
		Filters: req.Filters, SortBy: req.SortBy, SortOrder: req.SortOrder,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &authv1.ListRolesResponse{Total: total, Items: make([]*authv1.RoleResponse, len(result))}
	for i, role := range result {
		resp.Items[i] = toRoleResponse(&role)
	}
	return resp, nil
}

func (h *Handler) CreateRole(ctx context.Context, req *authv1.CreateRoleRequest) (*authv1.CreateRoleResponse, error) {
	if err := h.validate(req); err != nil {
		return nil, err
	}

	role, err := h.roleService.Create(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.CreateRoleResponse{Role: toRoleResponse(role)}, nil
}

func (h *Handler) UpdateRole(ctx context.Context, req *authv1.UpdateRoleRequest) (*authv1.UpdateRoleResponse, error) {
	if err := h.validate(req); err != nil {
		return nil, err
	}

	role, err := h.roleService.Update(ctx, req.Id, req.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "role")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.UpdateRoleResponse{Role: toRoleResponse(role)}, nil
}

func (h *Handler) DeleteRole(ctx context.Context, req *authv1.DeleteRoleRequest) (*authv1.DeleteRoleResponse, error) {
	if err := h.roleService.Delete(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.DeleteRoleResponse{Success: true}, nil
}

func toRoleResponse(role *roles.Role) *authv1.RoleResponse {
	if role == nil {
		return &authv1.RoleResponse{}
	}

	return &authv1.RoleResponse{
		Id:        uint64(role.ID),
		Name:      role.Name,
		CreatedAt: role.CreatedAt.Format(time.RFC3339),
		UpdatedAt: role.UpdatedAt.Format(time.RFC3339),
	}
}
