package service

import (
	"net/http"
	"pelaporan_prestasi/app/models/postgres"
	"pelaporan_prestasi/app/repository"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Definisi Role ID
const (
	RoleAdmin     = "11111111-1111-1111-1111-111111111111"
	RoleMahasiswa = "22222222-2222-2222-2222-222222222222"
	RoleDosen     = "33333333-3333-3333-3333-333333333333"
)

type UserService struct {
	UserRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{UserRepo: userRepo}
}

// --- STRUCT REQUEST BODY ---

type CreateUserRequest struct {
	Username string `json:"username" binding:"required" example:"mahasiswa_baru"`
	Email    string `json:"email" binding:"required,email" example:"mhs_baru@unair.ac.id"`
	Password string `json:"password" binding:"required,min=6" example:"rahasia123"`
	FullName string `json:"full_name" binding:"required" example:"Budi Santoso"`
	RoleID   string `json:"role_id" binding:"required" example:"22222222-2222-2222-2222-222222222222"`
}

type UpdateUserRequest struct {
	Email    string `json:"email" example:"email_baru@unair.ac.id"`
	FullName string `json:"full_name" example:"Budi Santoso Gelar Baru"`
	RoleID   string `json:"role_id" example:"22222222-2222-2222-2222-222222222222"`
}

type UpdateRoleRequest struct {
	RoleID string `json:"role_id" binding:"required" example:"33333333-3333-3333-3333-333333333333"`
}

// --- 1. GET ALL USERS (ADMIN ONLY) ---

// GetAllUsers godoc
// @Summary      Get All Users
// @Description  Hanya Admin.
// @Tags         Users (Admin)
// @Security     BearerAuth
// @Success      200  {object} map[string]interface{}
// @Failure      403  {object} map[string]string "Forbidden"
// @Router       /users [get]
func (s *UserService) GetAllUsers(c *gin.Context) {
	if c.GetString("role_id") != RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Hanya Admin yang boleh melihat daftar user."})
		return
	}
	users, err := s.UserRepo.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

// --- 2. GET USER BY ID (ADMIN & SELF) ---

// GetUserByID godoc
// @Summary      Get User Detail
// @Description  Admin bisa lihat siapa saja. User biasa hanya bisa lihat dirinya sendiri.
// @Tags         Users (Admin)
// @Security     BearerAuth
// @Param        id   path string true "User ID"
// @Success      200  {object} map[string]interface{}
// @Failure      403  {object} map[string]string "Forbidden"
// @Failure      404  {object} map[string]string "Not Found"
// @Router       /users/{id} [get]
func (s *UserService) GetUserByID(c *gin.Context) {
	myUserID := c.GetString("user_id")
	myRoleID := c.GetString("role_id")
	targetUserID := c.Param("id")

	// PROTEKSI: Jika BUKAN Admin DAN ID target BUKAN ID SAYA -> TOLAK
	if myRoleID != RoleAdmin {
		if myUserID != targetUserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak boleh melihat data user lain!"})
			return
		}
	}

	user, err := s.UserRepo.FindByID(c.Request.Context(), targetUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}
	// Kosongkan password hash biar aman
	user.PasswordHash = ""
	c.JSON(http.StatusOK, gin.H{"data": user})
}

// --- 3. CREATE USER (ADMIN ONLY) ---

// CreateUser godoc
// @Summary      Create User (Manual)
// @Description  Hanya Admin.
// @Tags         Users (Admin)
// @Security     BearerAuth
// @Param        request body CreateUserRequest true "Data User"
// @Success      201  {object} map[string]interface{}
// @Router       /users [post]
func (s *UserService) CreateUser(c *gin.Context) {
	if c.GetString("role_id") != RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak!"})
		return
	}

	var input CreateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	newUser := postgres.User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		FullName:     input.FullName,
		RoleID:       input.RoleID,
	}

	if err := s.UserRepo.Create(c.Request.Context(), &newUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": newUser})
}

// --- 4. UPDATE USER (ADMIN & SELF) ---

// UpdateUser godoc
// @Summary      Update User Data (Nama/Email)
// @Description  Admin bebas edit. User biasa hanya edit diri sendiri.
// @Tags         Users (Admin)
// @Security     BearerAuth
// @Param        id   path string true "User ID"
// @Param        request body UpdateUserRequest true "Data Update"
// @Success      200  {object} map[string]string
// @Router       /users/{id} [put]
func (s *UserService) UpdateUser(c *gin.Context) {
	myRoleID := c.GetString("role_id")
	myUserID := c.GetString("user_id")
	targetUserID := c.Param("id")

	if myRoleID != RoleAdmin {
		if myUserID != targetUserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Tidak boleh edit user lain!"})
			return
		}
	}

	var input UpdateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Proteksi: User biasa tidak boleh ganti RoleID lewat endpoint ini
	if myRoleID != RoleAdmin && input.RoleID != "" {
		input.RoleID = myRoleID
	}

	user := postgres.User{
		ID:       targetUserID,
		FullName: input.FullName,
		Email:    input.Email,
		RoleID:   input.RoleID,
	}

	if err := s.UserRepo.Update(c.Request.Context(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "User updated"})
}

// --- 5. UPDATE ROLE (ADMIN ONLY) ---

// UpdateRole godoc
// @Summary      Change User Role (Promote/Demote)
// @Description  Khusus Admin untuk mengubah jabatan user.
// @Tags         Users (Admin)
// @Security     BearerAuth
// @Param        id   path string true "User ID Target"
// @Param        request body UpdateRoleRequest true "Role ID Baru"
// @Success      200  {object} map[string]string
// @Failure      403  {object} map[string]string "Forbidden"
// @Router       /users/{id}/role [put]
func (s *UserService) UpdateRole(c *gin.Context) {
	// 1. Cek Admin
	if c.GetString("role_id") != RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Hanya Admin yang boleh ganti role!"})
		return
	}

	// 2. Ambil Input
	id := c.Param("id")
	var input UpdateRoleRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. Update ke Repo
	if err := s.UserRepo.UpdateRole(c.Request.Context(), id, input.RoleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Role berhasil diubah"})
}

// --- 6. DELETE USER (ADMIN ONLY) ---

// DeleteUser godoc
// @Summary      Soft Delete User
// @Description  Hanya Admin.
// @Tags         Users (Admin)
// @Security     BearerAuth
// @Param        id   path string true "User ID"
// @Success      200  {object} map[string]string
// @Router       /users/{id} [delete]
func (s *UserService) DeleteUser(c *gin.Context) {
	if c.GetString("role_id") != RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak!"})
		return
	}
	if err := s.UserRepo.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal delete"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "User deleted"})
}