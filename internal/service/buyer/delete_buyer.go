package buyer

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
)

func (s *buyerService) Delete(ctx context.Context, buyerID int64) error {
	existingBuyer, err := s.buyerRepo.GetBuyerByID(ctx, buyerID)
	if err != nil {
		return apperror.Internal("failed to get buyer before delete", err)
	}
	if existingBuyer == nil {
		return apperror.NotFound("buyer not found")
	}
	if err := s.buyerRepo.DeleteBuyer(ctx, buyerID); err != nil {
		return apperror.Internal("failed to delete buyer", err)
	}
	return nil
}
