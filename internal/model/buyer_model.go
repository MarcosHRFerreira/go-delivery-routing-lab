package model

import "time"

type BuyerModel struct {
	ID        int64
	Name      string
	Document  string
	Phone     string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
