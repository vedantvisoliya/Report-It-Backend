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

func setRefreshCookie(c *gin.Context, token string, maxAgeSeconds int) {
	c.SetCookie("refreshToken", token, maxAgeSeconds, "/", "", true, true)
}

// Refresh godoc
//
// @Summary      Rotate the token pair
// @Description  Reads the refresh token from the refreshToken httpOnly cookie, revokes it, and issues a fresh access token together with a new cookie.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  user.AuthResult
// @Failure      401  {object}  map[string]interface{}  "missing, revoked or expired refresh token"
// @Router       /auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refreshToken")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "missing refresh token",
			"ok": false,
		})
		return
	}

	output, err := h.svc.RefreshTokens(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
			"ok": false,
		})
		return 
	}

	setRefreshCookie(c, output.AccessToken, 14*24*60*60)
	c.JSON(http.StatusOK, output)
}

// Logout godoc
//
// @Summary      Log out
// @Description  Revokes the refresh token carried by the refreshToken cookie and clears that cookie. The /authorized/admin/auth/logout variant additionally requires a valid bearer token belonging to an admin.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "logged out successfully"
// @Failure      500  {object}  map[string]interface{}  "revoking the refresh token failed"
// @Router       /auth/logout [post]
// @Router       /authorized/admin/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie("refreshToken")
	if err := h.svc.Logout(c.Request.Context(), refreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"ok": false,
		})

		return
	}

	setRefreshCookie(c, "", -1)
	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
		"ok": true,
	})
}

// Register godoc
//
// @Summary      Register a new user
// @Description  Creates a Report It account. The email must be a college address (@glbitm.ac.in) and the password must be 8-16 characters long and contain at least one special character. On success the access token is returned in the body and the refresh token is set as an httpOnly cookie.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      user.RegisterInput  true  "Registration credentials"
// @Success      201      {object}  user.AuthResult
// @Failure      400      {object}  map[string]interface{}  "invalid JSON body, email already registered, or email/password validation failed"
// @Router       /auth/register [post]
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

	setRefreshCookie(c, output.RefreshToken, 14*24*60*60)
	c.JSON(http.StatusCreated, output)
}

// Login godoc
//
// @Summary      Log in
// @Description  Authenticates a user with email and password, returns an access token and sets the refresh token as an httpOnly cookie. The /authorized/admin/auth/login variant additionally requires a valid bearer token belonging to an admin.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      user.LoginInput  true  "Login credentials"
// @Success      200      {object}  user.AuthResult
// @Failure      400      {object}  map[string]interface{}  "invalid JSON body or invalid credentials"
// @Router       /auth/login [post]
// @Router       /authorized/admin/auth/login [post]
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

	setRefreshCookie(c, output.RefreshToken, 14*24*60*60)
	c.JSON(http.StatusOK, output)
}

// Update godoc
//
// @Summary      Update a user profile
// @Description  Updates the editable profile fields of the user identified by the path id. The payload is bound as a form, so send it as multipart/form-data or application/x-www-form-urlencoded.
// @Tags         users
// @Accept       mpfd
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        id          path      string  true   "User id (Mongo ObjectID hex)"
// @Param        profileUrl  formData  string  false  "Profile image URL"
// @Param        userName    formData  string  false  "Username"
// @Param        mobile      formData  string  false  "Mobile number"
// @Param        name        formData  string  false  "Full name"
// @Param        branch      formData  string  false  "Branch"
// @Param        year        formData  int     false  "Year of study"
// @Success      201         {object}  user.PublicUser
// @Failure      400         {object}  map[string]interface{}  "invalid form body or update failed"
// @Security     BearerAuth
// @Router       /user/{id} [put]
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

// GetAllUsers godoc
//
// @Summary      List users
// @Description  Returns a paginated list of users. page defaults to 1 and limit defaults to 10; the service clamps limit to the range 1-500.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        page   query     int  false  "Page number"  default(1)
// @Param        limit  query     int  false  "Page size"    default(10)
// @Success      200    {object}  user.PaginatedUser
// @Failure      400    {object}  map[string]interface{}  "page or limit is not a valid integer"
// @Failure      500    {object}  map[string]interface{}  "error fetching users"
// @Security     BearerAuth
// @Router       /user [get]
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

// GetSingleUserDetails godoc
//
// @Summary      Get a single user
// @Description  Returns the public profile of the user identified by the path id, wrapped in a data field.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User id (Mongo ObjectID hex)"
// @Success      200  {object}  map[string]interface{}  "data holds a user.PublicUser"
// @Failure      400  {object}  map[string]interface{}  "invalid user id or user not found"
// @Security     BearerAuth
// @Router       /user/{id} [get]
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

// DeleteUser godoc
//
// @Summary      Delete a user
// @Description  Deletes the user identified by the path id and removes their profile image from storage. Registered both as a user route and as an admin route.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User id (Mongo ObjectID hex)"
// @Success      200  {object}  map[string]interface{}  "message and the deleted user.PublicUser"
// @Failure      400  {object}  map[string]interface{}  "invalid user id, user not found, or profile image deletion failed"
// @Security     BearerAuth
// @Router       /user/{id} [delete]
// @Router       /authorized/admin/remove-user/{id} [delete]
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
