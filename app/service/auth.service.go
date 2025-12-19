package service

import (
	"net/http"
	"os"
	"strings"
	"time"

	"pelaporan_prestasi/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepo *repository.UserRepository
	JWTSecret string
}

func NewAuthService(userRepo *repository.UserRepository, secret string) *AuthService {
	return &AuthService{
		UserRepo:  userRepo,
		JWTSecret: secret, 
	}
}

// LoginRequest struct
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"mahasiswa123"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// --- 1. LOGIN ---

// Login godoc
// @Summary      Login User
// @Description  Masuk sistem dan dapatkan access token + refresh token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Kredensial Login"
// @Success      200  {object} map[string]interface{}
// @Failure      401  {object} map[string]string
// @Router       /auth/login [post]
func (s *AuthService) Login(c *gin.Context) {
	var input LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	user, err := s.UserRepo.FindByUsername(c.Request.Context(), input.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau password salah"})
		return
	}

	// PENGAMAN: Bersihkan spasi di database sebelum cek password
	cleanHash := strings.TrimSpace(user.PasswordHash)

	if err := bcrypt.CompareHashAndPassword([]byte(cleanHash), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau password salah"})
		return
	}

	// Generate Token
	accessToken, _ := s.generateToken(user.ID, user.RoleID, time.Hour*2)      // Expire 2 jam
	refreshToken, _ := s.generateToken(user.ID, user.RoleID, time.Hour*72)    // Expire 3 hari

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"token":        accessToken,
			"refreshToken": refreshToken,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"fullName": user.FullName,
				"role":     user.RoleID,
			},
		},
	})
}

// --- 2. REFRESH TOKEN ---

// RefreshToken godoc
// @Summary      Refresh Access Token
// @Description  Dapatkan token baru menggunakan refresh token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body object{refreshToken=string} true "Refresh Token"
// @Success      200  {object} map[string]interface{}
// @Failure      401  {object} map[string]string
// @Router       /auth/refresh [post]
func (s *AuthService) Refresh(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token required"})
		return
	}

	// Validasi Token
	claims, err := s.validateToken(input.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Generate Token Baru
	userID := claims["user_id"].(string)
	roleID := claims["role_id"].(string)
	newAccessToken, _ := s.generateToken(userID, roleID, time.Hour*2)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"token": newAccessToken,
		},
	})
}

// --- 3. LOGOUT ---

// Logout godoc
// @Summary      Logout User
// @Description  Keluar dari sistem
// @Tags         Authentication
// @Security     BearerAuth
// @Success      200  {object} map[string]string
// @Router       /auth/logout [post]
func (s *AuthService) Logout(c *gin.Context) {
	// Karena stateless, di backend hanya return success.
	// Di frontend nanti token dihapus dari localStorage.
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Logged out successfully"})
}

// --- 4. PROFILE ---

// GetProfile godoc
// @Summary      Get User Profile
// @Description  Mendapatkan data user yang sedang login (Butuh Token)
// @Tags         Authentication
// @Security     BearerAuth
// @Success      200  {object} map[string]interface{}
// @Router       /auth/profile [get]
func (s *AuthService) Profile(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
		return
	}
	
	// Hapus prefix "Bearer " jika ada
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	
	claims, err := s.validateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userID := claims["user_id"].(string)

	// Cari user di DB berdasarkan ID dari token
	user, err := s.UserRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": user,
	})
}

// --- HELPER FUNCTIONS ---

func (s *AuthService) generateToken(userID, roleID string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role_id": roleID,
		"exp":     time.Now().Add(duration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func (s *AuthService) validateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, os.ErrInvalid
}