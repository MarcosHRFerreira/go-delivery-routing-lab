package buyer

import (
	"context"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (s *buyerService) Create(ctx context.Context, req *dto.CreateBuyerRequest) (int64, error) {
	normalizedName, normalizedDocument, normalizedPhone, normalizedEmail := normalizeBuyerInput(
		req.Name,
		req.Document,
		req.Phone,
		req.Email,
	)

	now := time.Now()
	buyerModel := &model.BuyerModel{
		Name:      normalizedName,
		Document:  normalizedDocument,
		Phone:     normalizedPhone,
		Email:     normalizedEmail,
		CreatedAt: now,
		UpdatedAt: now,
	}

	buyerID, err := s.buyerRepo.CreateBuyer(ctx, buyerModel)
	if err != nil {
		if conflictErr := mapBuyerConflictError(err); conflictErr != nil {
			return 0, conflictErr
		}
		return 0, apperror.Internal("failed to create buyer", err)
	}

	return buyerID, nil
}
