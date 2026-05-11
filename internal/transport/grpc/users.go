package grpc

import (
	"context"
	"errors"
	"time"

	authv1 "auth-service/gen/auth/v1"
	"auth-service/internal/domain/users"
	"auth-service/internal/service"

	grpccrud "github.com/Mognus/go-grpc-crud/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func (h *Handler) GetUser(ctx context.Context, req *authv1.GetUserRequest) (*authv1.GetUserResponse, error) {
	user, err := h.userService.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.GetUserResponse{User: toUserResponse(&user)}, nil
}

func (h *Handler) ListUsers(ctx context.Context, req *authv1.ListUsersRequest) (*authv1.ListUsersResponse, error) {
	result, total, err := h.userService.List(ctx, grpccrud.ListRequest{
		Page: req.Page, Limit: req.Limit, Search: req.Search,
		Filters: req.Filters, SortBy: req.SortBy, SortOrder: req.SortOrder,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &authv1.ListUsersResponse{Total: total, Items: make([]*authv1.UserResponse, len(result))}
	for i, user := range result {
		resp.Items[i] = toUserResponse(&user)
	}
	return resp, nil
}

func (h *Handler) CreateUser(ctx context.Context, req *authv1.CreateUserRequest) (*authv1.CreateUserResponse, error) {
	if err := h.validate(req); err != nil {
		return nil, err
	}

	user, err := h.userService.Create(ctx, service.CreateUserInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		RoleID:    uint(req.RoleId),
		Active:    req.Active,
	})
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user with this email already exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.CreateUserResponse{User: toUserResponse(user)}, nil
}

func (h *Handler) UpdateUser(ctx context.Context, req *authv1.UpdateUserRequest) (*authv1.UpdateUserResponse, error) {
	if err := h.validate(req); err != nil {
		return nil, err
	}

	user, err := h.userService.Update(ctx, req.Id, service.UpdateUserInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		RoleID:    req.RoleId,
		Active:    req.Active,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.UpdateUserResponse{User: toUserResponse(user)}, nil
}

func (h *Handler) DeleteUser(ctx context.Context, req *authv1.DeleteUserRequest) (*authv1.DeleteUserResponse, error) {
	if err := h.userService.Delete(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.DeleteUserResponse{Success: true}, nil
}

func toUserResponse(user *users.User) *authv1.UserResponse {
	return &authv1.UserResponse{
		Id:        uint64(user.ID),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		RoleId:    uint64(user.RoleID),
		Role:      toRoleResponse(&user.Role),
		Active:    user.Active != nil && *user.Active,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}
