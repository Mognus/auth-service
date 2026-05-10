package schema

type Field struct {
	Name         string         `json:"name"`
	Type         string         `json:"type"`
	Label        string         `json:"label"`
	Required     bool           `json:"required,omitempty"`
	Readonly     bool           `json:"readonly,omitempty"`
	TableHidden  bool           `json:"tableHidden,omitempty"`
	EditHidden   bool           `json:"editHidden,omitempty"`
	CreateHidden bool           `json:"createHidden,omitempty"`
	Options      []SelectOption `json:"options,omitempty"`
}

type SelectOption struct {
	Value any    `json:"value"`
	Label string `json:"label"`
}

type Schema struct {
	Name        string   `json:"name"`
	APIPath     string   `json:"apiPath"`
	DisplayName string   `json:"displayName"`
	Fields      []Field  `json:"fields"`
	Searchable  []string `json:"searchable"`
}

var Users = Schema{
	Name:        "users",
	APIPath:     "auth/users",
	DisplayName: "Users",
	Fields: []Field{
		{Name: "id", Type: "number", Label: "ID", Readonly: true, EditHidden: true, CreateHidden: true},
		{Name: "email", Type: "string", Label: "Email", Required: true},
		{Name: "password", Type: "string", Label: "Password", Required: true, EditHidden: true},
		{Name: "firstName", Type: "string", Label: "First Name"},
		{Name: "lastName", Type: "string", Label: "Last Name"},
		{Name: "roleId", Type: "relation", Label: "Role", Required: true, TableHidden: true},
		{Name: "role", Type: "object", Label: "Role", Readonly: true, EditHidden: true, TableHidden: true, CreateHidden: true},
		{Name: "active", Type: "boolean", Label: "Active"},
		{Name: "createdAt", Type: "date", Label: "Created", Readonly: true, EditHidden: true, CreateHidden: true},
		{Name: "updatedAt", Type: "date", Label: "Updated", Readonly: true, EditHidden: true, CreateHidden: true},
	},
	Searchable: []string{"id", "email", "firstName", "lastName", "active"},
}

var Roles = Schema{
	Name:        "roles",
	APIPath:     "auth/roles",
	DisplayName: "Roles",
	Fields: []Field{
		{Name: "id", Type: "number", Label: "ID", Readonly: true},
		{Name: "name", Type: "string", Label: "Name", Required: true},
		{Name: "createdAt", Type: "date", Label: "Created", Readonly: true},
		{Name: "updatedAt", Type: "date", Label: "Updated", Readonly: true},
	},
	Searchable: []string{"name"},
}
