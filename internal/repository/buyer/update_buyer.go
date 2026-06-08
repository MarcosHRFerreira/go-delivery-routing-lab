package buyer

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *buyerRepository) UpdateBuyer(ctx context.Context, model *model.BuyerModel) error {
	query := `UPDATE buyers
	SET name = ?,
	document = ?,
	phone = ?,
	email =?,
	updated_at = ?
	WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query,
		model.Name,
		model.Document,
		model.Phone,
		model.Email,
		model.UpdatedAt,
		model.ID,
	)
	if err != nil {
		return err
	}
	return nil
}
