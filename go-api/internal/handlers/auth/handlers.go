package handlers_auth

import (
	"encoding/json"
	"errors"
	"go-api/internal/helpers"
	"go-api/internal/logger"
	"go-api/internal/service/tokens"
	types_external "go-api/internal/types/external"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetToken godoc
// @Summary Generate authentication token
// @Description Get a token for authentication (auto-creates user if not exists)
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body types_external.GenerateTokenRequest true "Credentials"
// @Success 200 {object} map[string]interface{}
// @Router /auth/getToken [post]
func GetToken(c *gin.Context) {
	var req types_external.GenerateTokenRequest

	if c.Request.Body == nil {
		helpers.Respond(c, http.StatusBadRequest, "empty_body", "request body is empty")
		return
	}

	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			helpers.Respond(c, http.StatusBadRequest, "empty_body", "request body is empty")
			return
		}

		var unmarshalTypeError *json.UnmarshalTypeError
		if errors.As(err, &unmarshalTypeError) {
			helpers.Respond(c, http.StatusBadRequest, "invalid_type", err.Error())
			return
		}

		if strings.Contains(err.Error(), "unknown field") {
			helpers.Respond(c, http.StatusBadRequest, "unknown_field", err.Error())
			return
		}

		helpers.Respond(c, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	if req.Password == "" || req.User == "" {
		helpers.Respond(c, http.StatusBadRequest, "validation_error", "FIELD: user and password are required")
		return
	}

	token, err := tokens.GenerateToken(req.User, req.TelegramID)
	if err != nil {
		logger.Error().Err(err).Str("username", req.User).Msg("failed to generate token")
		helpers.Respond(c, http.StatusBadRequest, "db_error", err.Error())
		return
	}

	logger.Info().Str("username", req.User).Msg("token generated")
	c.JSON(200, gin.H{"token": token})
}
