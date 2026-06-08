package buyer

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

type BuyerRepository interface {
	CreateBuyer(ctx context.Context, model *model.BuyerModel) (int64, error)
	GetBuyerByID(ctx context.Context, buyerID int64) (*model.BuyerModel, error)
	GetAllBuyers(ctx context.Context) ([]model.BuyerModel, error)
	UpdateBuyer(ctx context.Context, model *model.BuyerModel) error
	DeleteBuyer(ctx context.Context, buyerID int64) error
}
type buyerRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) BuyerRepository {
	return &buyerRepository{
		db: db,
	}
}
