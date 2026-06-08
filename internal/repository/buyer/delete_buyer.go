package buyer

import "context"

func (r *buyerRepository) DeleteBuyer(ctx context.Context, buyerID int64) error {
	query := `DELETE FROM buyers WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, buyerID)
	if err != nil {
		return err
	}
	return nil
}
