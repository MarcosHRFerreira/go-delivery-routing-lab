package buyer

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *buyerRepository) CreateBuyer(ctx context.Context, model *model.BuyerModel) (int64, error) {

	query := `INSERT INTO buyers (name, document, phone, email, created_at, updated_at)
	VALUES (?,?,?,?,?,?)`

	result, err := r.db.ExecContext(
		ctx,
		query,
		model.Name,
		model.Document,
		model.Phone,
		model.Email,
		model.CreatedAt,
		model.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	buyerID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return buyerID, nil

}
