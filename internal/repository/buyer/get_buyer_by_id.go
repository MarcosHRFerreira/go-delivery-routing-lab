package buyer

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *buyerRepository) GetBuyerByID(ctx context.Context, buyerID int64) (*model.BuyerModel, error) {
	query := `SELECT id, name, document, phone, email, created_at, updated_at
	FROM buyers WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, buyerID)

	var result model.BuyerModel
	err := row.Scan(
		&result.ID,
		&result.Name,
		&result.Document,
		&result.Phone,
		&result.Email,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil

}
