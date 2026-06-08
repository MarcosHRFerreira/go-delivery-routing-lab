package dto

import "time"

type (
	CreateBuyerRequest struct {
		Name     string `json:"name" validate:"required,min=3"`
		Document string `json:"document" validate:"required"`
		Phone    string `json:"phone" validate:"required"`
		Email    string `json:"email" validate:"omitempty,email"`
	}
	UpdateBuyerRequest struct {
		Name     string `json:"name" validate:"required,min=3"`
		Document string `json:"document" validate:"required"`
		Phone    string `json:"phone" validate:"required"`
		Email    string `json:"email" validate:"omitempty,email"`
	}
	BuyerResponse struct {
		ID        int64     `json:"id"`
		Name      string    `json:"name"`
		Document  string    `json:"document"`
		Phone     string    `json:"phone"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
)
