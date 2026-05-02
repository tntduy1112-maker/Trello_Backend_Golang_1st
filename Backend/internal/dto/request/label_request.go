package request

type CreateLabelRequest struct {
	Name  *string `json:"name" validate:"omitempty,max=100"`
	Color string  `json:"color" validate:"required,hexcolor"`
}

type UpdateLabelRequest struct {
	Name  *string `json:"name" validate:"omitempty,max=100"`
	Color *string `json:"color" validate:"omitempty,hexcolor"`
}
