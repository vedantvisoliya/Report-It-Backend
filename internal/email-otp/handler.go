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
