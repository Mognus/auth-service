package client

import (
	"context"
	"sync"

	authv1 "auth-service/gen/auth/v1"

	fiberhandler "github.com/Mognus/go-grpc-crud"
	libcrud "github.com/Mognus/go-grpc-crud/crud"
	apperrors "github.com/Mognus/go-grpc-crud/errors"

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
			Searchable: []string{"email", "firstName", "lastName"},
		}
	})
	return p.schema
}

func (p *UserProvider) HandleSchema(c *fiber.Ctx) error { return c.JSON(p.GetSchema()) }

func (p *UserProvider) HandleList(c *fiber.Ctx) error {
	params := fiberhandler.ParseListParams(c)

	resp, err := p.grpcClient.ListUsers(c.UserContext(), &authv1.ListUsersRequest{
		Page: params.Page, Limit: params.Limit, Search: params.Search,
		Filters: params.Filters, SortBy: params.SortBy, SortOrder: params.SortOrder,
	})
	if err != nil {
		return apperrors.GrpcToHTTP(err)
	}

	items := make([]any, len(resp.Users))
	for i, u := range resp.Users {
		items[i] = ToUserJSON(u)
	}

	return fiberhandler.WriteList(c, items, resp.Total, params.Page, params.Limit)
}

func (p *UserProvider) HandleGet(c *fiber.Ctx) error {
	id, err := fiberhandler.ParseID(c)
	if err != nil {
		return err
	}

	resp, err := p.grpcClient.GetUser(c.UserContext(), &authv1.GetUserRequest{Id: id})
	if err != nil {
		return apperrors.GrpcToHTTP(err)
	}

	return c.JSON(ToUserJSON(resp.User))
}

func (p *UserProvider) HandleCreate(c *fiber.Ctx) error {
	req := &authv1.CreateUserRequest{}
	if err := fiberhandler.DecodeBody(c, req); err != nil {
		return err
	}

	resp, err := p.grpcClient.CreateUser(c.UserContext(), req)
	if err != nil {
		return apperrors.GrpcToHTTP(err)
	}

	return fiberhandler.WriteCreated(c, ToUserJSON(resp.User))
}

func (p *UserProvider) HandleUpdate(c *fiber.Ctx) error {
	id, err := fiberhandler.ParseID(c)
	if err != nil {
		return err
	}

	req := &authv1.UpdateUserRequest{}
	if err := fiberhandler.DecodeBody(c, req); err != nil {
		return err
	}

	req.Id = id
	resp, err := p.grpcClient.UpdateUser(c.UserContext(), req)
	if err != nil {
		return apperrors.GrpcToHTTP(err)
	}

	return c.JSON(ToUserJSON(resp.User))
}

func (p *UserProvider) HandleDelete(c *fiber.Ctx) error {
	id, err := fiberhandler.ParseID(c)
	if err != nil {
		return err
	}

	if _, err := p.grpcClient.DeleteUser(c.UserContext(), &authv1.DeleteUserRequest{Id: id}); err != nil {
		return apperrors.GrpcToHTTP(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
