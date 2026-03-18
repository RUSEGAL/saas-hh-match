package payments

import (
	dbpayments "go-api/internal/repository/payments"
	types_external "go-api/internal/types/external"
	types_internal "go-api/internal/types/int"
)

func CreatePayment(id int64, amount int64, status string, provider string) (user_id int64, err error) {
	return dbpayments.AddPaymentDb(id, amount, status, provider)
}
func UpdatePayment(id, userID int64, req *types_external.PaymentRequest) error {
	return dbpayments.UpdatePaymentDb(
		id,
		userID,
		req.Amount,
		req.Status,
		req.Provider,
	)
}
func GetPaymentByUser(userID int64) ([]types_internal.Payment, error) {
	return dbpayments.GetPaymentsByUserDb(userID)
}
func GetAllPayments() {

}

func ProcessYookassaPayment(userID int64, paymentID string, amount int64, status, provider string) error {
	_, err := dbpayments.AddPaymentDb(userID, amount, status, provider)
	return err
}
