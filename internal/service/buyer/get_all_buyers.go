package buyer

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
)

func (s *buyerService) GetAll(ctx context.Context) ([]dto.BuyerResponse, error) {
	buyers, err := s.buyerRepo.GetAllBuyers(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to get buyers", err)
	}

	response := make([]dto.BuyerResponse, 0, len(buyers))
	for _, buyer := range buyers {
		response = append(response, dto.BuyerResponse{
			ID:        buyer.ID,
			Name:      buyer.Name,
			Document:  buyer.Document,
			Phone:     buyer.Phone,
			Email:     buyer.Email,
			CreatedAt: buyer.CreatedAt,
		})
	}
	return response, nil
}
