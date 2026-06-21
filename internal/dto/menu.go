package modelDto

type UpdateMenuRequest struct {
	Label              *string `json:"label"`
	Icon               *string `json:"icon"`
	RequiredPermission *string `json:"required_permission"` // Deprecated
	SortOrder          *int    `json:"sort_order"`
	IsSystem           *bool   `json:"is_system"`
	AllowedRoles       []uint  `json:"allowed_roles"`
}

type CreateMenuRequest struct {
	Key                string `json:"key"`
	Label              string `json:"label" binding:"required"`
	Icon               string `json:"icon" binding:"required"`
	Path               string `json:"path"`
	SortOrder          int    `json:"sort_order"`
	IsSystem           bool   `json:"is_system"`
	RequiredPermission string `json:"required_permission"` // Deprecated
	ParentID           *uint  `json:"parent_id"`
	AllowedRoles       []uint `json:"allowed_roles"`
}
