package grpc

import (
	"context"
	"fmt"

	authv1 "auth-service/gen/auth/v1"

	grpccrud "github.com/Mognus/go-grpc-crud/server"
)

func (h *Handler) GetUsersSchema(ctx context.Context, _ *authv1.GetUsersSchemaRequest) (*authv1.GetUsersSchemaResponse, error) {
	roles, _, _ := h.roleService.List(ctx, grpccrud.ListRequest{Limit: 100})
	roleOptions := make([]*authv1.SchemaSelectOption, len(roles))
	for i, r := range roles {
		roleOptions[i] = &authv1.SchemaSelectOption{
			Value: fmt.Sprintf("%d", r.ID),
			Label: r.Name,
		}
	}

	return &authv1.GetUsersSchemaResponse{
		Name:        "users",
		DisplayName: "Users",
		Fields: []*authv1.SchemaField{
			{Name: "id", Type: "number", Label: "ID", Readonly: true, EditHidden: true, CreateHidden: true},
			{Name: "email", Type: "string", Label: "Email", Required: true},
			{Name: "password", Type: "string", Label: "Password", Required: true, EditHidden: true},
			{Name: "firstName", Type: "string", Label: "First Name"},
			{Name: "lastName", Type: "string", Label: "Last Name"},
			{Name: "roleId", Type: "relation", Label: "Role", Required: true, TableHidden: true, Options: roleOptions},
			{Name: "role", Type: "object", Label: "Role", Readonly: true, EditHidden: true, TableHidden: true, CreateHidden: true},
			{Name: "active", Type: "boolean", Label: "Active"},
			{Name: "createdAt", Type: "date", Label: "Created", Readonly: true, EditHidden: true, CreateHidden: true},
			{Name: "updatedAt", Type: "date", Label: "Updated", Readonly: true, EditHidden: true, CreateHidden: true},
		},
		Searchable: []string{"id", "email", "firstName", "lastName", "active"},
	}, nil
}

func (h *Handler) GetRolesSchema(_ context.Context, _ *authv1.GetRolesSchemaRequest) (*authv1.GetRolesSchemaResponse, error) {
	return &authv1.GetRolesSchemaResponse{
		Name:        "roles",
		DisplayName: "Roles",
		Fields: []*authv1.SchemaField{
			{Name: "id", Type: "number", Label: "ID", Readonly: true},
			{Name: "name", Type: "string", Label: "Name", Required: true},
			{Name: "createdAt", Type: "date", Label: "Created", Readonly: true},
			{Name: "updatedAt", Type: "date", Label: "Updated", Readonly: true},
		},
		Searchable: []string{"name"},
	}, nil
}
