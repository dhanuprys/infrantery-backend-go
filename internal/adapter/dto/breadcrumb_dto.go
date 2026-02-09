package dto

type BreadcrumbItem struct {
	Type     string           `json:"type"`               // "project", "diagram", "note", "node", "vault"
	ID       string           `json:"id,omitempty"`       // ID of the resource
	Label    string           `json:"label"`              // Name or title for display
	Active   bool             `json:"active"`             // Whether this is the current active item
	Siblings []BreadcrumbItem `json:"siblings,omitempty"` // Siblings at the same level (e.g., other diagrams in the same folder)
}

type BreadcrumbResponse struct {
	ProjectID string           `json:"project_id"`
	Path      []BreadcrumbItem `json:"path"`
}
