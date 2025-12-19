package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Ambil Header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort() // Stop! Jangan lanjut ke controller
			return
		}

		// 2. Format harus "Bearer <token>"
		// Kita buang kata "Bearer " untuk ambil kode token-nya saja
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		// 3. Parsing & Validasi Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Pastikan metode enkripsinya sesuai (HMAC)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// Ambil Secret Key dari .env
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		// 4. Jika Error atau Token Tidak Valid
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// 5. SUKSES! Simpan data user ke Context
		// Supaya nanti di Controller kita tahu siapa yang sedang login
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID := claims["user_id"].(string)
			roleID := claims["role_id"].(string)

			// Set variable ini biar bisa dipanggil pake c.Get("user_id") nanti
			c.Set("user_id", userID)
			c.Set("role_id", roleID)
		}

		// Lanjut ke Controller (GetAllUsers, CreateUser, dll)
		c.Next()
	}
}