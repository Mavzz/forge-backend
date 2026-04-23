package models

type Category struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"` // Useful for frontend UI branding
	Tasks []Task `json:"tasks"`
}
