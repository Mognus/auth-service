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

var userListConfig = grpccrud.ListConfig{
	Preloads:        []string{"Role"},
	Searchable:      []string{"email", "first_name", "last_name"},
	SortableColumns: []string{"id", "email", "first_name", "last_name", "created_at", "updated_at", "active"},
	DefaultSort:     "id ASC",
}

func (h *Handler) GetUser(ctx context.Context, req *authv1.GetUserRequest) (*authv1.GetUserResponse, error) {
	user, err := grpccrud.DefaultGet[model.User](ctx, h.db, req.Id, "Role")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.GetUserResponse{User: toUserResponse(&user)}, nil
}

func (h *Handler) ListUsers(ctx context.Context, req *authv1.ListUsersRequest) (*authv1.ListUsersResponse, error) {
	users, total, err := grpccrud.DefaultList[model.User](ctx, h.db, grpccrud.ListRequest{
		Page: req.Page, Limit: req.Limit, Search: req.Search,
		Filters: req.Filters, SortBy: req.SortBy, SortOrder: req.SortOrder,
	}, userListConfig)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := &authv1.ListUsersResponse{Total: total, Users: make([]*authv1.UserResponse, len(users))}
	for i, u := range users {
		resp.Users[i] = toUserResponse(&u)
	}
	return resp, nil
}

func (h *Handler) CreateUser(ctx context.Context, req *authv1.CreateUserRequest) (*authv1.CreateUserResponse, error) {
	user, err := grpccrud.DefaultCreate(ctx, h.db, &model.User{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		RoleID:    uint(req.RoleId),
		Active:    req.Active,
	}, "Role")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.CreateUserResponse{User: toUserResponse(user)}, nil
}

func (h *Handler) UpdateUser(ctx context.Context, req *authv1.UpdateUserRequest) (*authv1.UpdateUserResponse, error) {
	updates := map[string]any{
		"email":      req.Email,
		"first_name": req.FirstName,
		"last_name":  req.LastName,
		"role_id":    req.RoleId,
		"active":     req.Active,
	}
	if req.Password != "" {
		updates["password"] = req.Password
	}
	user, err := grpccrud.DefaultUpdate[model.User](ctx, h.db, req.Id, updates, "Role")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.UpdateUserResponse{User: toUserResponse(user)}, nil
}

func (h *Handler) DeleteUser(ctx context.Context, req *authv1.DeleteUserRequest) (*authv1.DeleteUserResponse, error) {
	if err := grpccrud.DefaultDelete(ctx, h.db, &model.User{}, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.DeleteUserResponse{Success: true}, nil
}

func toUserResponse(u *model.User) *authv1.UserResponse {
	return &authv1.UserResponse{
		Id:        uint64(u.ID),
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		RoleId:    uint64(u.RoleID),
		Role:      toRoleResponse(&u.Role),
		Active:    u.Active,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}
}
