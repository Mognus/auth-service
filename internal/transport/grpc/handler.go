package grpc

import (
	"errors"
	"log"
	"strings"
	"time"

	authv1 "auth-service/gen/auth/v1"
	"auth-service/internal/service"

	protovalidate "buf.build/go/protovalidate"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

type Handler struct {
	authv1.UnimplementedAuthServiceServer
	validator   protovalidate.Validator
	authService *service.AuthService
	roleService *service.RoleService
	userService *service.UserService
}

func New(db *gorm.DB, jwtSecret string, accessTTL, refreshTTL time.Duration) *Handler {
	v, err := protovalidate.New()
	if err != nil {
		log.Fatal("failed to initialize validator:", err)
	}

	return &Handler{
		validator:   v,
		authService: service.NewAuthService(db, jwtSecret, accessTTL, refreshTTL),
		roleService: service.NewRoleService(db),
		userService: service.NewUserService(db),
	}
}

func (h *Handler) validate(msg proto.Message) error {
	if err := h.validator.Validate(msg); err != nil {
		var valErr *protovalidate.ValidationError
		if errors.As(err, &valErr) && len(valErr.Violations) > 0 {
			parts := make([]string, 0, len(valErr.Violations))
			for _, v := range valErr.Violations {
				parts = append(parts, v.String())
			}
			return status.Error(codes.InvalidArgument, strings.Join(parts, ", "))
		}
		return status.Error(codes.InvalidArgument, err.Error())
	}

	return nil
}
