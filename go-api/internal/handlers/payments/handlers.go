package payments_handlers

import (
	"go-api/internal/helpers"
	"go-api/internal/logger"
	"go-api/internal/service/payments"
	types_external "go-api/internal/types/external"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreatePayment godoc
// @Summary Create a new payment
// @Description Create a new payment for the authenticated user
// @Tags Payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body types_external.PaymentRequest true "Payment data"
// @Success 200 {object} map[string]interface{}
// @Router /api/payments [post]
func CreatePayment(c *gin.Context) {
	var req types_external.PaymentRequest
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.Respond(c, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.Respond(c, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if req.Amount <= 0 {
		helpers.Respond(c, http.StatusBadRequest, "invalid_amount", "amount must be > 0")
		return
	}
	id, err := payments.CreatePayment(userID, req.Amount, req.Status, req.Provider)
	if err != nil {
		logger.Error().Err(err).Int64("user_id", userID).Msg("failed to create payment")
		helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	logger.Info().Int64("user_id", userID).Int64("payment_id", id).Int64("amount", req.Amount).Msg("payment created")
	c.JSON(200, gin.H{"id": id})
}

// UpdatePayment godoc
// @Summary Update a payment
// @Description Update an existing payment by ID
// @Tags Payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Payment ID"
// @Param request body types_external.PaymentRequest true "Payment data"
// @Success 200 {object} map[string]interface{}
// @Router /api/payments/{id} [patch]
func UpdatePayment(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.Respond(c, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		helpers.Respond(c, http.StatusBadRequest, "invalid_id", "wrong id format")
		return
	}

	var req types_external.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.Respond(c, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	err = payments.UpdatePayment(id, userID, &req)
	if err != nil {
		logger.Error().Err(err).Int64("payment_id", id).Int64("user_id", userID).Msg("failed to update payment")
		helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	logger.Info().Int64("payment_id", id).Int64("user_id", userID).Msg("payment updated")
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// GetMyPayments godoc
// @Summary Get my payments
// @Description Get all payments for the authenticated user
// @Tags Payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []types_internal.Payment
// @Router /api/payments/me [get]
func GetMyPayments(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.Respond(c, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	userPayments, err := payments.GetPaymentByUser(userID)
	if err != nil {
		logger.Error().Err(err).Int64("user_id", userID).Msg("failed to get payments")
		helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	logger.Info().Int64("user_id", userID).Int("count", len(userPayments)).Msg("payments retrieved")
	c.JSON(http.StatusOK, gin.H{"payments": userPayments})
}
