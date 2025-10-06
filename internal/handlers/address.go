package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/henrique/address-validator/internal/models"
	"github.com/henrique/address-validator/internal/services"
)

type AddressHandler struct {
	validatorService *services.ValidatorService
}

func NewAddressHandler(validatorService *services.ValidatorService) *AddressHandler {
	return &AddressHandler{
		validatorService: validatorService,
	}
}

// ValidateAddress godoc
// @Summary      Validate and normalize an address
// @Description  Receives a free-form address, corrects typos automatically and returns the normalized components
// @Tags         address
// @Accept       json
// @Produce      json
// @Param        address  body      models.ValidateAddressRequest  true  "Address to validate"
// @Success      200      {object}  models.ValidateAddressResponse
// @Failure      400      {object}  models.ValidateAddressResponse
// @Failure      401      {object}  map[string]string "Unauthorized - Invalid or missing token"
// @Failure      415      {object}  map[string]string "Unsupported Media Type - Content-Type must be application/json"
// @Failure      500      {object}  models.ValidateAddressResponse
// @Security     BearerAuth
// @Router       /validate-address [post]
func (h *AddressHandler) ValidateAddress(c *gin.Context) {
	var req models.ValidateAddressRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ValidateAddressResponse{
			Status: "error",
			Error:  "Invalid request: address field is required",
		})
		return
	}

	result, err := h.validatorService.ValidateAddress(c.Request.Context(), req.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ValidateAddressResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
