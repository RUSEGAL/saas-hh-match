package dbpayments

import (
	"go-api/internal/config/db"
	types_internal "go-api/internal/types/int"
)

func AddPaymentDb(userID int64, amount int64, status, provider string) (int64, error) {
	var id int64
	err := db.DB.QueryRow(`
		INSERT INTO payments (user_id, amount, status, provider)
		VALUES ($1, $2, $3, $4) RETURNING id
	`, userID, amount, status, provider).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}
func UpdatePaymentDb(id, userID int64, amount int64, status, provider string) error {
	_, err := db.DB.Exec(`
		UPDATE payments
		SET amount = $1, status = $2, provider = $3
		WHERE id = $4 AND user_id = $5
	`, amount, status, provider, id, userID)

	return err
}
func GetPaymentsByUserDb(userID int64) ([]types_internal.Payment, error) {
	rows, err := db.DB.Query(`
		SELECT id, user_id, amount, status, provider, created_at
		FROM payments
		WHERE user_id = $1
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []types_internal.Payment

	for rows.Next() {
		var p types_internal.Payment

		err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.Amount,
			&p.Status,
			&p.Provider,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		payments = append(payments, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return payments, nil
}
