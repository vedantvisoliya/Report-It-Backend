package post

import (
	"net/http"
	"reportit-api/internal/middleware"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewPostHandler(svc *Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) CreatePost(c *gin.Context) {
	userID := c.Param("id")
	var form CreatePostForm

	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"ok":    false,
		})
		return
	}

	post, err := h.svc.CreatePost(c.Request.Context(), form, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "post created successfully",
		"data":    post,
	})
}

func (h *Handler) GetAllUserPosts(c *gin.Context) {
	userID := c.Param("id")

	page, err := strconv.ParseInt(
		c.DefaultQuery("Page", "1"),
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

	data, err := h.svc.GetAllPosts(c.Request.Context(), page, limit, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "error fetching posts",
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (h *Handler) GetSinglePost(c *gin.Context) {
	postID := c.Param("id")

	data, err := h.svc.GetSinglePost(c.Request.Context(), postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (h *Handler) RemovePost(c *gin.Context) {
	postID := c.Param("id")
	role, ok := middleware.GetUserRole(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "error getting role",
			"ok":    false,
		})
		return
	}

	var userID DeletePostRequest
	if err := c.ShouldBindJSON(&userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"ok":    false,
		})
		return
	}

	post, err := h.svc.DeletePost(c.Request.Context(), postID, userID.UserID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "post deleted successfully",
		"post":    post,
	})
}

func (h *Handler) UpdatePost(c *gin.Context) {
	postID := c.Param("id")
	var input UpdatePostForm

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid JSON body",
			"ok":    false,
		})
		return
	}

	update, err := h.svc.UpdatePost(c.Request.Context(), postID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, update)
}
