package buyer

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
)

func (s *buyerService) GetByID(ctx context.Context, buyerID int64) (*dto.BuyerResponse, error) {
	buyerModel, err := s.buyerRepo.GetBuyerByID(ctx, buyerID)
	if err != nil {
		return nil, apperror.Internal("failed to get buyer", err)
	}
	if buyerModel == nil {
		return nil, apperror.NotFound("buyer not found")
	}

	return &dto.BuyerResponse{
		ID:        buyerModel.ID,
		Name:      buyerModel.Name,
		Document:  buyerModel.Document,
		Phone:     buyerModel.Phone,
		Email:     buyerModel.Email,
		CreatedAt: buyerModel.CreatedAt,
		UpdatedAt: buyerModel.UpdatedAt,
	}, nil
}
