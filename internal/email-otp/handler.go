package emailotp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewOTPHandler(svc *Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

// SendOTP godoc
//
// @Summary      Send an email OTP
// @Description  Generates a one time password for the given email address and delivers it over email.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      emailotp.SendReq  true  "Target email address"
// @Success      200      {object}  map[string]interface{}  "OTP send successfully"
// @Failure      400      {object}  map[string]interface{}  "invalid JSON body or missing email"
// @Failure      500      {object}  map[string]interface{}  "generating or sending the OTP failed"
// @Router       /auth/send-otp [post]
func (h *Handler) SendOTP(c *gin.Context) {
	var sendReq SendReq
	if err := c.ShouldBindJSON(&sendReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"ok":    false,
		})
		return
	}

	err := h.svc.SendOTP(c.Request.Context(), sendReq.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP send successfully",
		"ok":      true,
	})
}

// VerifyOTP godoc
//
// @Summary      Verify an email OTP
// @Description  Verifies the one time password previously sent to the given email address.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      emailotp.VerifyReq  true  "Email address and OTP"
// @Success      202      {object}  map[string]interface{}  "OTP verified successfully"
// @Failure      400      {object}  map[string]interface{}  "invalid JSON body, or expired or incorrect OTP"
// @Router       /auth/verify-otp [post]
func (h *Handler) VerifyOTP(c *gin.Context) {
	var verifyReq VerifyReq
	if err := c.ShouldBindJSON(&verifyReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid JSON body",
			"ok":      false,
		})
		return
	}

	err := h.svc.VerifyOTP(c.Request.Context(), verifyReq.Email, verifyReq.OTP)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"ok":      false,
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "OTP verified successfully",
		"ok":      true,
	})
}
