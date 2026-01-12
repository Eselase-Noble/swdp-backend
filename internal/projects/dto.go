package projects

type ProjectResponse struct {
	ProjectID   string `json:"project_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProjectName string `json:"project_name" example:"SWDP Backend"`
	OwnerID     string `json:"owner_id" example:"f47ac10b-58cc-4372-a567-0e02b2c3d479"`
}
