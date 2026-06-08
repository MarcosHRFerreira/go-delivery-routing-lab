package buyer

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	buyerrepo "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/buyer"
)

type BuyerService interface {
	Create(ctx context.Context, req *dto.CreateBuyerRequest) (int64, error)
	GetByID(ctx context.Context, buyerID int64) (*dto.BuyerResponse, error)
	GetAll(ctx context.Context) ([]dto.BuyerResponse, error)
	Update(ctx context.Context, buyerID int64, req *dto.UpdateBuyerRequest) (*dto.BuyerResponse, error)
	Delete(ctx context.Context, buyerID int64) error
}

type buyerService struct {
	buyerRepo buyerrepo.BuyerRepository
}

func NewService(buyerRepo buyerrepo.BuyerRepository) BuyerService {
	return &buyerService{
		buyerRepo: buyerRepo,
	}
}
