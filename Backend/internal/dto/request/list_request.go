package request

type CreateListRequest struct {
	Title string `json:"title" validate:"required,min=1,max=255"`
}

type UpdateListRequest struct {
	Title string `json:"title" validate:"required,min=1,max=255"`
}

type MoveListRequest struct {
	Position float64 `json:"position" validate:"required,gt=0"`
}

type CopyListRequest struct {
	Title string `json:"title" validate:"required,min=1,max=255"`
}
