package buyer

import (
	"context"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (s *buyerService) Update(ctx context.Context, buyerID int64, req *dto.UpdateBuyerRequest) (*dto.BuyerResponse, error) {
	existingBuyer, err := s.buyerRepo.GetBuyerByID(ctx, buyerID)
	if err != nil {
		return nil, apperror.Internal("failed to get buyer before update", err)
	}
	if existingBuyer == nil {
		return nil, apperror.NotFound("buyer not found")
	}

	normalizedName, normalizedDocument, normalizedPhone, normalizedEmail := normalizeBuyerInput(
		req.Name,
		req.Document,
		req.Phone,
		req.Email,
	)

	existingBuyer.Name = normalizedName
	existingBuyer.Document = normalizedDocument
	existingBuyer.Phone = normalizedPhone
	existingBuyer.Email = normalizedEmail
	existingBuyer.UpdatedAt = time.Now()

	if err := s.buyerRepo.UpdateBuyer(ctx, &model.BuyerModel{
		ID:        existingBuyer.ID,
		Name:      existingBuyer.Name,
		Document:  existingBuyer.Document,
		Phone:     existingBuyer.Phone,
		Email:     existingBuyer.Email,
		CreatedAt: existingBuyer.CreatedAt,
		UpdatedAt: existingBuyer.UpdatedAt,
	}); err != nil {
		if conflictErr := mapBuyerConflictError(err); conflictErr != nil {
			return nil, conflictErr
		}
		return nil, apperror.Internal("failed to update buyer", err)
	}

	return &dto.BuyerResponse{
		ID:        existingBuyer.ID,
		Name:      existingBuyer.Name,
		Document:  existingBuyer.Document,
		Phone:     existingBuyer.Phone,
		Email:     existingBuyer.Email,
		CreatedAt: existingBuyer.CreatedAt,
		UpdatedAt: existingBuyer.UpdatedAt,
	}, nil
}
