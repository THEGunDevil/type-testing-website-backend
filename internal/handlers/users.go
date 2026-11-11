package handlers

import (
	"log"
	"net/http"
	"strings"
	"typetype/internal/database"
	"typetype/internal/models"
	"typetype/internal/utils/jwt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	DB         *database.Queries // Receives the universal query wrapper
	JWTManager *jwt.JWTManager
}
type RegistrationRequest struct {
	UserName string `json:"user_name" binding:"required,min=6,max=8"`
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,min=8,max=12"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var req RegistrationRequest

	// 1. Bind and Validate JSON payload
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Printf("[Registration Failed] Input validation error: %v | Client IP: %s", err, c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Hash the password safely using Bcrypt
	// Use bcrypt.DefaultCost (10) for a secure balance between speed and security
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[Registration Error] Failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// 3. Save to Database (Mocking database logic here)
	user, err := h.DB.CreateUser(c.Request.Context(), database.CreateUserParams{
		UserName:     req.UserName,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		log.Printf("[DB Error] Failed saving user: %v", err)
		c.JSON(http.StatusConflict, gin.H{"error": "Username or Email already registered"})
		return
	}
	// 4. Send success response (Never return the hashedPassword back to the client)
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"user_name": user.UserName,
			"email":     user.Email,
		},
	})
}

type LogInRequest struct {
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,min=8,max=12"`
}

func (h *UserHandler) LogIn(c *gin.Context) {
	var req LogInRequest

	// 1. Bind and Validate JSON payload
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Printf("[Login Failed] Input validation error: %v | Client IP: %s", err, c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Verify requested password with Hashed password safely using Bcrypt
	user, err := h.DB.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}
	accessToken, err := h.JWTManager.GenerateAccessToken(user.ID.String(), user.Role, user.TokenVersion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := h.JWTManager.GenerateRefreshToken(user.ID.String(), user.Role, user.TokenVersion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// Determine if running on localhost based on Host header
	isLocalhost := strings.Contains(c.Request.Host, "localhost") || strings.Contains(c.Request.Host, "127.0.0.1")
	accessCookie := &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   15 * 60, // 15 minutes
		HttpOnly: true,
		Secure:   !isLocalhost,
		SameSite: http.SameSiteLaxMode,
	}

	if !isLocalhost {
		accessCookie.SameSite = http.SameSiteNoneMode
	}

	http.SetCookie(c.Writer, accessCookie)
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",                  // Ensure cookie is available site-wide
		MaxAge:   3600 * 24 * 7,        // 7 days
		HttpOnly: true,                 // Prevent JavaScript access
		Secure:   !isLocalhost,         // Secure=true in production, false on localhost
		SameSite: http.SameSiteLaxMode, // Default to Lax for compatibility
	}

	// Use SameSite=None for cross-origin requests in production
	if !isLocalhost {
		cookie.SameSite = http.SameSiteNoneMode
	}

	// Avoid setting Domain explicitly unless necessary
	// If backend is on a different domain (e.g., api.himel-s-library.vercel.app), uncomment and set:
	// cookie.Domain = "your-backend-domain.com"

	http.SetCookie(c.Writer, cookie)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": models.UserResponse{
			UserName:  user.UserName,
			Role: user.Role,
			Email:     user.Email,
			AvatarUrl: user.AvatarUrl.String,
		},
	})
}
func (h *UserHandler) Refresh(c *gin.Context) {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing refresh token"})
		return
	}

	claims, err := h.JWTManager.VerifyRefreshToken(cookie)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}
	userIDStr := claims.ID
	version := claims.TokenVersion

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.DB.GetUserByID(c, pgtype.UUID{Bytes: userUUID, Valid: true})
	if err != nil || user.TokenVersion != int32(version) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired or invalid"})
		return
	}

	accessToken, err := h.JWTManager.GenerateAccessToken(userIDStr, user.Role, user.TokenVersion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new access token"})
		return
	}

	refreshToken, err := h.JWTManager.GenerateRefreshToken(userIDStr, user.Role, user.TokenVersion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	isLocalhost := strings.Contains(c.Request.Host, "localhost") || strings.Contains(c.Request.Host, "127.0.0.1")

	// 1. Set the new access_token cookie
	accessCookie := &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   15 * 60, // 15 minutes
		HttpOnly: true,
		Secure:   !isLocalhost,
		SameSite: http.SameSiteLaxMode,
	}
	if !isLocalhost {
		accessCookie.SameSite = http.SameSiteNoneMode
	}
	http.SetCookie(c.Writer, accessCookie)

	// 2. Set the new refresh_token cookie
	refreshTokenCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   3600 * 24 * 7, // 7 days
		HttpOnly: true,
		Secure:   !isLocalhost,
		SameSite: http.SameSiteLaxMode,
	}
	if !isLocalhost {
		refreshTokenCookie.SameSite = http.SameSiteNoneMode
	}
	http.SetCookie(c.Writer, refreshTokenCookie)

	// 3. Respond with success JSON (cookies handle the tokens automatically)
	c.JSON(http.StatusOK, gin.H{"message": "Token refreshed successfully"})
}

// LogoutHandler clears the refresh token cookie and invalidates token version
func (h *UserHandler) LogOut(c *gin.Context) {
	// Determine if running on localhost
	isLocalhost := strings.Contains(c.Request.Host, "localhost") || strings.Contains(c.Request.Host, "127.0.0.1")
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Expire immediately
		HttpOnly: true,
		Secure:   !isLocalhost,
		SameSite: http.SameSiteLaxMode,
	}

	if !isLocalhost {
		cookie.SameSite = http.SameSiteNoneMode
	}

	// Invalidate token version for security
	cookieValue, err := c.Cookie("refresh_token")
	if err == nil {
		claims, err := h.JWTManager.VerifyRefreshToken(cookieValue)
		if err == nil {
			userIDStr := claims.ID
			userUUID, err := uuid.Parse(userIDStr)
			if err != nil {
				log.Printf("Failed to parse user id: %v", err)
			}
			if _, err := h.DB.IncrementTokenVersion(c.Request.Context(), pgtype.UUID{Bytes: userUUID, Valid: true}); err != nil {
				// Log error but don't fail logout
				log.Printf("Failed to increment token version: %v", err)
			}
		}
	}

	http.SetCookie(c.Writer, cookie)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.GetString("userID")

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	user, err := h.DB.GetUserByID(
		c.Request.Context(),
		pgtype.UUID{
			Bytes: userUUID,
			Valid: true,
		},
	)
	// In your AuthMiddleware or GetUser handler:
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user": models.UserResponse{
			UserName:  user.UserName,
			Role: user.Role,
			Email:     user.Email,
			AvatarUrl: user.AvatarUrl.String,
		},
	})
}
