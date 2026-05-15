package modelDto

type UpdateMenuRequest struct {
	Label        *string  `json:"label"`
	Icon         *string  `json:"icon"`
	AllowedRoles []string `json:"allowed_roles"`
	SortOrder    *int     `json:"sort_order"`
	IsSystem     *bool    `json:"is_system"`
}
