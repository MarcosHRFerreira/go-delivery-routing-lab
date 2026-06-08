package buyer

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *buyerRepository) GetAllBuyers(ctx context.Context) ([]model.BuyerModel, error) {

	query := `SELECT id, name, document, phone, email, created_at, updated_at
	FROM buyers
	ORDER BY id ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]model.BuyerModel, 0)
	for rows.Next() {
		var buyer model.BuyerModel
		if err := rows.Scan(
			&buyer.ID,
			&buyer.Name,
			&buyer.Document,
			&buyer.Phone,
			&buyer.Email,
			&buyer.CreatedAt,
			&buyer.UpdatedAt,
		); err != nil {
			return nil, err
		}

		results = append(results, buyer)

	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil

}
