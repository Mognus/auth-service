package provider

import (
	"context"

	authv1 "auth-service/gen/auth/v1"
	client "auth-service/client"
	libcrud "github.com/Mognus/go-grpc-crud/crud"
	grpccrud "github.com/Mognus/go-grpc-crud/proxy"

	"github.com/gofiber/fiber/v2"
)

type RoleProvider struct {
	grpcClient authv1.AuthServiceClient
}

func NewRoleProvider(grpcClient authv1.AuthServiceClient) *RoleProvider {
	return &RoleProvider{grpcClient: grpcClient}
}

func (p *RoleProvider) GetModelName() string { return "roles" }

func (p *RoleProvider) GetSchema() libcrud.Schema {
	return libcrud.Schema{
		Name:        "roles",
		DisplayName: "Roles",
		Fields: []libcrud.Field{
			{Name: "id", Type: "number", Label: "ID", Readonly: true},
			{Name: "name", Type: "string", Label: "Name", Required: true},
			{Name: "createdAt", Type: "date", Label: "Created", Readonly: true},
			{Name: "updatedAt", Type: "date", Label: "Updated", Readonly: true},
		},
		Searchable: []string{"name"},
	}
}

func (p *RoleProvider) SchemaHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(p.GetSchema())
	}
}

func (p *RoleProvider) RegisterRoutes(router fiber.Router) {
	roles := router.Group("/roles")
	roles.Get("/", p.ListHandler())
	roles.Get("/schema", p.SchemaHandler())
	roles.Get("/:id", p.GetHandler())
	roles.Post("/", p.CreateHandler())
	roles.Put("/:id", p.UpdateHandler())
	roles.Delete("/:id", p.DeleteHandler())
}

func (p *RoleProvider) ListHandler() fiber.Handler {
	return grpccrud.DefaultListProxy(func(ctx context.Context, page, limit int32, search string, filters map[string]string, sortBy, sortOrder string) ([]any, int64, error) {
		resp, err := p.grpcClient.ListRoles(ctx, &authv1.ListRolesRequest{
			Page: page, Limit: limit, Search: search,
			Filters: filters, SortBy: sortBy, SortOrder: sortOrder,
		})
		if err != nil {
			return nil, 0, err
		}
		items := make([]any, len(resp.Roles))
		for i, r := range resp.Roles {
			v := client.ToRoleJSON(r)
			items[i] = &v
		}
		return items, resp.Total, nil
	})
}

func (p *RoleProvider) GetHandler() fiber.Handler {
	return grpccrud.DefaultGetProxy(func(ctx context.Context, id uint64) (any, error) {
		resp, err := p.grpcClient.GetRole(ctx, &authv1.GetRoleRequest{Id: id})
		if err != nil {
			return nil, err
		}
		return client.ToRoleJSON(resp.Role), nil
	})
}

func (p *RoleProvider) CreateHandler() fiber.Handler {
	return grpccrud.DefaultCreateProxy(func(ctx context.Context, req *authv1.CreateRoleRequest) (any, error) {
		resp, err := p.grpcClient.CreateRole(ctx, req)
		if err != nil {
			return nil, err
		}
		return client.ToRoleJSON(resp.Role), nil
	})
}

func (p *RoleProvider) UpdateHandler() fiber.Handler {
	return grpccrud.DefaultUpdateProxy(func(ctx context.Context, id uint64, req *authv1.UpdateRoleRequest) (any, error) {
		req.Id = id
		resp, err := p.grpcClient.UpdateRole(ctx, req)
		if err != nil {
			return nil, err
		}
		return client.ToRoleJSON(resp.Role), nil
	})
}

func (p *RoleProvider) DeleteHandler() fiber.Handler {
	return grpccrud.DefaultDeleteProxy(func(ctx context.Context, id uint64) error {
		_, err := p.grpcClient.DeleteRole(ctx, &authv1.DeleteRoleRequest{Id: id})
		return err
	})
}
