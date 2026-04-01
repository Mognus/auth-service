package client

import authv1 "auth-service/gen/auth/v1"

type UserJSON struct {
	ID        uint64    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	RoleID    uint64    `json:"roleId"`
	Role      *RoleJSON `json:"role"`
	Active    bool      `json:"active"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
}

type RoleJSON struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func ToUserJSON(u *authv1.UserResponse) *UserJSON {
	if u == nil {
		return nil
	}
	return &UserJSON{
		ID:        u.Id,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		RoleID:    u.RoleId,
		Role:      ToRoleJSON(u.Role),
		Active:    u.Active,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func ToRoleJSON(r *authv1.RoleResponse) *RoleJSON {
	if r == nil {
		return &RoleJSON{}
	}
	return &RoleJSON{
		ID:        r.Id,
		Name:      r.Name,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
