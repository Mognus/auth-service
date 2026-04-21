package client

import (
	authv1 "auth-service/gen/auth/v1"

	fiberhandler "github.com/Mognus/go-grpc-crud"
	libcrud "github.com/Mognus/go-grpc-crud/crud"
	apperrors "github.com/Mognus/go-grpc-crud/errors"

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

func (p *RoleProvider) HandleSchema(c *fiber.Ctx) error { return c.JSON(p.GetSchema()) }

func (p *RoleProvider) HandleList(c *fiber.Ctx) error {
	params := fiberhandler.ParseListParams(c)

	resp, err := p.grpcClient.ListRoles(c.UserContext(), &authv1.ListRolesRequest{
		Page: params.Page, Limit: params.Limit, Search: params.Search,
		Filters: params.Filters, SortBy: params.SortBy, SortOrder: params.SortOrder,
	})
	if err != nil {
		return apperrors.GrpcToHTTP(err)
	}

	items := make([]any, len(resp.Roles))
	for i, r := range resp.Roles {
		items[i] = ToRoleJSON(r)
	}

	return fiberhandler.WriteList(c, items, resp.Total, params.Page, params.Limit)
}

func (p *RoleProvider) HandleGet(c *fiber.Ctx) error {
	id, err := fiberhandler.ParseID(c)
	if err != nil {
		return err
	}

	resp, err := p.grpcClient.GetRole(c.UserContext(), &authv1.GetRoleRequest{Id: id})
	if err != nil {
		return apperrors.GrpcToHTTP(err)
	}

	return c.JSON(ToRoleJSON(resp.Role))
}

func (p *RoleProvider) HandleCreate(c *fiber.Ctx) error {
	req := &authv1.CreateRoleRequest{}
	if err := fiberhandler.DecodeBody(c, req); err != nil {
		return err
	}

	resp, err := p.grpcClient.CreateRole(c.UserContext(), req)
	if err != nil {
		return apperrors.GrpcToHTTP(err)
	}

	return fiberhandler.WriteCreated(c, ToRoleJSON(resp.Role))
}

func (p *RoleProvider) HandleUpdate(c *fiber.Ctx) error {
	id, err := fiberhandler.ParseID(c)
	if err != nil {
		return err
	}

	req := &authv1.UpdateRoleRequest{}
	if err := fiberhandler.DecodeBody(c, req); err != nil {
		return err
	}

	req.Id = id
	resp, err := p.grpcClient.UpdateRole(c.UserContext(), req)
	if err != nil {
		return apperrors.GrpcToHTTP(err)
	}

	return c.JSON(ToRoleJSON(resp.Role))
}

func (p *RoleProvider) HandleDelete(c *fiber.Ctx) error {
	id, err := fiberhandler.ParseID(c)
	if err != nil {
		return err
	}

	if _, err := p.grpcClient.DeleteRole(c.UserContext(), &authv1.DeleteRoleRequest{Id: id}); err != nil {
		return apperrors.GrpcToHTTP(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
