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

// CreatePost godoc
//
// @Summary      Create a report post
// @Description  Creates a report (lost and found, harassment, ragging, canteen overcharging, and so on) authored by the user in the path id. The payload is bound as a form: title (max 100 characters), description (max 5000 characters) and postType are required, and an optional image is validated, resized, compressed and uploaded to storage.
// @Tags         posts
// @Accept       mpfd
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        id           path      string  true   "Author user id (Mongo ObjectID hex)"
// @Param        title        formData  string  true   "Post title (max 100 characters)"
// @Param        description  formData  string  true   "Post description (max 5000 characters)"
// @Param        postType     formData  string  true   "Report category"
// @Param        isAnonymous  formData  bool    false  "Publish the report anonymously"
// @Param        image        formData  file    false  "Optional image attachment"
// @Success      200          {object}  map[string]interface{}  "message and the created post.Post"
// @Failure      400          {object}  map[string]interface{}  "invalid form body, validation failure, invalid image, upload failure or invalid user id"
// @Security     BearerAuth
// @Router       /post/{id} [post]
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

	// // TEMP DEBUG — remove after diagnosing
	// if err := c.Request.ParseMultipartForm(32 << 20); err == nil {
	// 	for key, values := range c.Request.MultipartForm.Value {
	// 		fmt.Printf("RAW FORM KEY: %q VALUES: %v\n", key, values)
	// 	}
	// }

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

// GetAllPosts godoc
//
// @Summary      List all posts
// @Description  Returns a paginated list of every report post. The page query parameter is read as Page (capital P) and defaults to 1; limit defaults to 10 and is clamped to the range 1-100.
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        Page   query     int  false  "Page number"  default(1)
// @Param        limit  query     int  false  "Page size"    default(10)
// @Success      200    {object}  post.PaginatedAllPosts
// @Failure      400    {object}  map[string]interface{}  "Page or limit is not a valid integer, or error fetching posts"
// @Security     BearerAuth
// @Router       /post [get]
func (h *Handler) GetAllPosts(c *gin.Context) {
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

	data, err := h.svc.GetAllPosts(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "error fetching posts",
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetAllUserPosts godoc
//
// @Summary      List the posts of one user
// @Description  Returns a paginated list of the posts authored by the user in the path id, filtered by the anonymous flag. The page query parameter is read as Page (capital P) and defaults to 1; limit defaults to 10 and is clamped to the range 1-100.
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        id         path      string  true   "Author user id (Mongo ObjectID hex)"
// @Param        anonymous  query     bool    false  "Return anonymous posts"  default(false)
// @Param        Page       query     int     false  "Page number"             default(1)
// @Param        limit      query     int     false  "Page size"               default(10)
// @Success      200        {object}  post.PaginatedAllPosts
// @Failure      400        {object}  map[string]interface{}  "anonymous is not a boolean, Page or limit is not a valid integer, or error fetching posts"
// @Security     BearerAuth
// @Router       /post/user/{id} [get]
func (h *Handler) GetAllUserPosts(c *gin.Context) {
	userID := c.Param("id")

	isAnonymous, err := strconv.ParseBool(
		c.DefaultQuery("anonymous", "false"),
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "anonymous must be boolean value",
			"ok":    false,
		})
		return
	}

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

	data, err := h.svc.GetAllUserPosts(c.Request.Context(), page, limit, userID, isAnonymous)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "error fetching posts",
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetSinglePost godoc
//
// @Summary      Get a single post
// @Description  Returns the report post identified by the path id.
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Post id (Mongo ObjectID hex)"
// @Success      200  {object}  post.Post
// @Failure      400  {object}  map[string]interface{}  "invalid post id or post not found"
// @Security     BearerAuth
// @Router       /post/{id} [get]
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

// RemovePost godoc
//
// @Summary      Delete a post
// @Description  Deletes the report post identified by the path id. The caller role is taken from the access token and the owning user id is taken from the request body. Registered both as a user route and as an admin route.
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        id       path      string                  true  "Post id (Mongo ObjectID hex)"
// @Param        request  body      post.DeletePostRequest  true  "Owning user id"
// @Success      200      {object}  map[string]interface{}  "message and the deleted post.Post"
// @Failure      400      {object}  map[string]interface{}  "role missing from the context, invalid JSON body, or deletion failed"
// @Security     BearerAuth
// @Router       /post/{id} [delete]
// @Router       /authorized/admin/remove-post/{id} [delete]
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

// UpdatePost godoc
//
// @Summary      Update a post
// @Description  Updates the report post identified by the path id. The payload is bound as a form, so send it as multipart/form-data or application/x-www-form-urlencoded. postType is required, and an optional replacement image is validated, resized, compressed and uploaded to storage.
// @Tags         posts
// @Accept       mpfd
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        id           path      string  true   "Post id (Mongo ObjectID hex)"
// @Param        title        formData  string  false  "Post title (max 100 characters)"
// @Param        description  formData  string  false  "Post description (max 5000 characters)"
// @Param        postType     formData  string  true   "Report category"
// @Param        image        formData  file    false  "Replacement image attachment"
// @Success      200          {object}  post.Post
// @Failure      400          {object}  map[string]interface{}  "invalid form body, validation failure, invalid image, upload failure or update failed"
// @Security     BearerAuth
// @Router       /post/{id} [patch]
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
