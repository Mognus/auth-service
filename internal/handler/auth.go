package handler

import (
	"context"
	"errors"
	"time"

	authv1 "auth-service/gen/auth/v1"
	"auth-service/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Handler struct {
	authv1.UnimplementedAuthServiceServer
	db        *gorm.DB
	jwtSecret string
}

func New(db *gorm.DB, jwtSecret string) *Handler {
	return &Handler{db: db, jwtSecret: jwtSecret}
}

// --- Auth ---

func (h *Handler) Login(_ context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	var user model.User
	if err := h.db.Preload("Role").Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}
	if !user.Active {
		return nil, status.Error(codes.PermissionDenied, "account is deactivated")
	}
	if !user.CheckPassword(req.Password) {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}
	token, err := h.generateToken(&user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}
	return &authv1.LoginResponse{Token: token, User: toUserResponse(&user)}, nil
}

func (h *Handler) Register(_ context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	var existing model.User
	if err := h.db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return nil, status.Error(codes.AlreadyExists, "user with this email already exists")
	}
	if len(req.Password) < 8 {
		return nil, status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	}

	var defaultRole model.Role
	if err := h.db.Where("name = ?", string(model.RoleUser)).First(&defaultRole).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to find default role")
	}

	user := model.User{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		RoleID:    defaultRole.ID,
		Active:    true,
	}
	if err := h.db.Create(&user).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to create user")
	}
	user.Role = defaultRole

	token, err := h.generateToken(&user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}
	return &authv1.RegisterResponse{Token: token, User: toUserResponse(&user)}, nil
}

// --- Users ---

func (h *Handler) GetUser(_ context.Context, req *authv1.GetUserRequest) (*authv1.GetUserResponse, error) {
	var user model.User
	if err := h.db.Preload("Role").First(&user, req.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.GetUserResponse{User: toUserResponse(&user)}, nil
}

func (h *Handler) ListUsers(_ context.Context, req *authv1.ListUsersRequest) (*authv1.ListUsersResponse, error) {
	page, limit := int(req.Page), int(req.Limit)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	query := h.db.Model(&model.User{}).Preload("Role")
	if req.Search != "" {
		s := "%" + req.Search + "%"
		query = query.Where("email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?", s, s, s)
	}
	if req.RoleId > 0 {
		query = query.Where("role_id = ?", req.RoleId)
	}
	if req.Active != nil {
		query = query.Where("active = ?", *req.Active)
	}

	var total int64
	query.Count(&total)

	var users []model.User
	if err := query.Offset((page - 1) * limit).Limit(limit).Find(&users).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &authv1.ListUsersResponse{Total: total}
	for i := range users {
		resp.Users = append(resp.Users, toUserResponse(&users[i]))
	}
	return resp, nil
}

func (h *Handler) CreateUser(_ context.Context, req *authv1.CreateUserRequest) (*authv1.CreateUserResponse, error) {
	user := model.User{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		RoleID:    uint(req.RoleId),
		Active:    req.Active,
	}
	if err := h.db.Create(&user).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	h.db.Preload("Role").First(&user, user.ID)
	return &authv1.CreateUserResponse{User: toUserResponse(&user)}, nil
}

func (h *Handler) UpdateUser(_ context.Context, req *authv1.UpdateUserRequest) (*authv1.UpdateUserResponse, error) {
	var user model.User
	if err := h.db.First(&user, req.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

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

	if err := h.db.Model(&user).Updates(updates).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	h.db.Preload("Role").First(&user, user.ID)
	return &authv1.UpdateUserResponse{User: toUserResponse(&user)}, nil
}

func (h *Handler) DeleteUser(_ context.Context, req *authv1.DeleteUserRequest) (*authv1.DeleteUserResponse, error) {
	if err := h.db.Delete(&model.User{}, req.Id).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.DeleteUserResponse{Success: true}, nil
}

// --- Roles ---

func (h *Handler) GetRole(_ context.Context, req *authv1.GetRoleRequest) (*authv1.GetRoleResponse, error) {
	var role model.Role
	if err := h.db.First(&role, req.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "role not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.GetRoleResponse{Role: toRoleResponse(&role)}, nil
}

func (h *Handler) ListRoles(_ context.Context, req *authv1.ListRolesRequest) (*authv1.ListRolesResponse, error) {
	page, limit := int(req.Page), int(req.Limit)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}

	query := h.db.Model(&model.Role{})
	if req.Search != "" {
		query = query.Where("name ILIKE ?", "%"+req.Search+"%")
	}

	var total int64
	query.Count(&total)

	var roles []model.Role
	if err := query.Offset((page - 1) * limit).Limit(limit).Find(&roles).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &authv1.ListRolesResponse{Total: total}
	for i := range roles {
		resp.Roles = append(resp.Roles, toRoleResponse(&roles[i]))
	}
	return resp, nil
}

func (h *Handler) CreateRole(_ context.Context, req *authv1.CreateRoleRequest) (*authv1.CreateRoleResponse, error) {
	role := model.Role{Name: req.Name}
	if err := h.db.Create(&role).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.CreateRoleResponse{Role: toRoleResponse(&role)}, nil
}

func (h *Handler) UpdateRole(_ context.Context, req *authv1.UpdateRoleRequest) (*authv1.UpdateRoleResponse, error) {
	var role model.Role
	if err := h.db.First(&role, req.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "role not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := h.db.Model(&role).Update("name", req.Name).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.UpdateRoleResponse{Role: toRoleResponse(&role)}, nil
}

func (h *Handler) DeleteRole(_ context.Context, req *authv1.DeleteRoleRequest) (*authv1.DeleteRoleResponse, error) {
	if err := h.db.Delete(&model.Role{}, req.Id).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.DeleteRoleResponse{Success: true}, nil
}

// --- Helpers ---

func (h *Handler) generateToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role.Name,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
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

func toRoleResponse(r *model.Role) *authv1.RoleResponse {
	return &authv1.RoleResponse{
		Id:        uint64(r.ID),
		Name:      r.Name,
		CreatedAt: r.CreatedAt.Format(time.RFC3339),
		UpdatedAt: r.UpdatedAt.Format(time.RFC3339),
	}
}
