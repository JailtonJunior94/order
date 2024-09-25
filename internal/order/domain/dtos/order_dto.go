package dtos

type (
	OrderInput struct {
		Items []*OrderItemInput `json:"items"`
	}

	OrderItemInput struct {
		ProductName string  `json:"product_name"`
		Price       float64 `json:"price"`
		Quantity    uint    `json:"quantity"`
	}

	OrderOutput struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
)

func NewOrderOutput(id string, status string) *OrderOutput {
	return &OrderOutput{
		ID:     id,
		Status: status,
	}
}
