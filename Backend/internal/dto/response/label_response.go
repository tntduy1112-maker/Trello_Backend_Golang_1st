package response

type LabelResponse struct {
	ID    string  `json:"id"`
	Name  *string `json:"name"`
	Color string  `json:"color"`
}

type LabelSummary struct {
	ID    string  `json:"id"`
	Name  *string `json:"name"`
	Color string  `json:"color"`
}
