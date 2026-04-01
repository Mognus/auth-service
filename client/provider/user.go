package provider

import (
	"context"
	"sync"

	authv1 "auth-service/gen/auth/v1"
	client "auth-service/client"
	libcrud "github.com/Mognus/go-grpc-crud/crud"
	grpccrud "github.com/Mognus/go-grpc-crud/proxy"

	"github.com/gofiber/fiber/v2"
)

type UserProvider struct {
	grpcClient authv1.AuthServiceClient
	schema     libcrud.Schema
	schemaOnce sync.Once
}

func NewUserProvider(grpcClient authv1.AuthServiceClient) *UserProvider {
	return &UserProvider{grpcClient: grpcClient}
}

func (p *UserProvider) GetModelName() string { return "users" }

func (p *UserProvider) GetSchema() libcrud.Schema {
	p.schemaOnce.Do(func() {
		rolesResp, _ := p.grpcClient.ListRoles(context.Background(), &authv1.ListRolesRequest{Limit: 100})
		var roleOptions []libcrud.SelectOption
		if rolesResp != nil {
			for _, r := range rolesResp.Roles {
				roleOptions = append(roleOptions, libcrud.SelectOption{Value: r.Id, Label: r.Name})
			}
		}
		p.schema = libcrud.Schema{
			Name:        "users",
			DisplayName: "Users",
			Fields: []libcrud.Field{
				{Name: "id", Type: "number", Label: "ID", Readonly: true},
				{Name: "email", Type: "string", Label: "Email", Required: true},
				{Name: "password", Type: "string", Label: "Password", Required: true, EditHidden: true},
				{Name: "firstName", Type: "string", Label: "First Name"},
				{Name: "lastName", Type: "string", Label: "Last Name"},
				{Name: "roleId", Type: "relation", Label: "Role", Required: true, TableHidden: true, Options: roleOptions},
				{Name: "role", Type: "object", Label: "Role", Readonly: true},
				{Name: "active", Type: "boolean", Label: "Active"},
				{Name: "createdAt", Type: "date", Label: "Created", Readonly: true},
				{Name: "updatedAt", Type: "date", Label: "Updated", Readonly: true},
			},
			Searchable: []string{"email", "first_name", "last_name"},
		}
	})
	return p.schema
}

func (p *UserProvider) SchemaHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(p.GetSchema())
	}
}


func (p *UserProvider) ListHandler() fiber.Handler {
	return grpccrud.DefaultListProxy(func(ctx context.Context, page, limit int32, search string, filters map[string]string, sortBy, sortOrder string) ([]any, int64, error) {
		resp, err := p.grpcClient.ListUsers(ctx, &authv1.ListUsersRequest{
			Page: page, Limit: limit, Search: search,
			Filters: filters, SortBy: sortBy, SortOrder: sortOrder,
		})
		if err != nil {
			return nil, 0, err
		}
		items := make([]any, len(resp.Users))
		for i, u := range resp.Users {
			items[i] = client.ToUserJSON(u)
		}
		return items, resp.Total, nil
	})
}

func (p *UserProvider) GetHandler() fiber.Handler {
	return grpccrud.DefaultGetProxy(func(ctx context.Context, id uint64) (any, error) {
		resp, err := p.grpcClient.GetUser(ctx, &authv1.GetUserRequest{Id: id})
		if err != nil {
			return nil, err
		}
		return client.ToUserJSON(resp.User), nil
	})
}

func (p *UserProvider) CreateHandler() fiber.Handler {
	return grpccrud.DefaultCreateProxy(func(ctx context.Context, req *authv1.CreateUserRequest) (any, error) {
		resp, err := p.grpcClient.CreateUser(ctx, req)
		if err != nil {
			return nil, err
		}
		return client.ToUserJSON(resp.User), nil
	})
}

func (p *UserProvider) UpdateHandler() fiber.Handler {
	return grpccrud.DefaultUpdateProxy(func(ctx context.Context, id uint64, req *authv1.UpdateUserRequest) (any, error) {
		req.Id = id
		resp, err := p.grpcClient.UpdateUser(ctx, req)
		if err != nil {
			return nil, err
		}
		return client.ToUserJSON(resp.User), nil
	})
}

func (p *UserProvider) DeleteHandler() fiber.Handler {
	return grpccrud.DefaultDeleteProxy(func(ctx context.Context, id uint64) error {
		_, err := p.grpcClient.DeleteUser(ctx, &authv1.DeleteUserRequest{Id: id})
		return err
	})
}
