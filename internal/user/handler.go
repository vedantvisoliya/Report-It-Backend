package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewUserHandler(svc *Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid JSON body",
			"ok":    false,
		})
		return
	}

	output, err := h.svc.Register(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusCreated, output)
}

func (h *Handler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid JSON body",
			"ok":    false,
		})
		return
	}

	output, err := h.svc.Login(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, output)
}

func (h *Handler) Update(c *gin.Context) {
	var input UpdateUserForm
	userID := c.Param("id")

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid JSON body",
			"ok":    false,
		})
		return
	}

	user, err := h.svc.UpdateUser(c.Request.Context(), input, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *Handler) GetAllUsers(c *gin.Context) {
	page, err := strconv.ParseInt(
		c.DefaultQuery("page", "1"),
		10,
		64,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "page must be a valid integer",
			"ok":    false,
		})
		return
	}

	limit, err := strconv.ParseInt(
		c.DefaultQuery("limit", "10"),
		10,
		64,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "limit must be a valid integer",
			"ok":    false,
		})
		return
	}

	data, err := h.svc.GetAllUsers(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error fetching users",
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (h *Handler) GetSingleUserDetails(c *gin.Context) {
	userID := c.Param("id")
	user, err := h.svc.GetSingleUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	user, err := h.svc.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
		"user":    user,
	})
}
