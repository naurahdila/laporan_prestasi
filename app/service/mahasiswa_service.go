package service

import (
	"pelaporan_prestasi/app/models/dto"
	"pelaporan_prestasi/app/repository"

	"github.com/gin-gonic/gin"
)

type UpdateAdvisorRequest struct {
	AdvisorID string `json:"advisor_id" binding:"required"`
}

type MahasiswaService struct {
	Repo    *repository.MahasiswaRepository
	AchRepo *repository.AchievementRepository
}

func NewMahasiswaService(repo *repository.MahasiswaRepository, achRepo *repository.AchievementRepository) *MahasiswaService {
	return &MahasiswaService{Repo: repo, AchRepo: achRepo}
}

// GetAll godoc
// @Summary Get All Mahasiswa (Filtered by Role)
// @Tags Mahasiswa
// @Security BearerAuth
// @Router /mahasiswa [get]
func (s *MahasiswaService) GetAll(c *gin.Context) {
	loginUserID := c.GetString("user_id")
	loginRoleID := c.GetString("role_id")

	data, err := s.Repo.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal mengambil data: " + err.Error()})
		return
	}

	var filteredData []repository.MahasiswaData
	for _, m := range data {
		if loginRoleID == RoleAdmin {
			filteredData = append(filteredData, m)
		} else if loginRoleID == RoleDosen && m.AdvisorUserID == loginUserID {
			filteredData = append(filteredData, m)
		} else if loginRoleID == RoleMahasiswa && m.UserID == loginUserID {
			filteredData = append(filteredData, m)
		}
	}

	if filteredData == nil {
		filteredData = []repository.MahasiswaData{}
	}
	c.JSON(200, gin.H{"data": filteredData})
}

// GetDetail godoc
// @Summary Get Mahasiswa Detail (Privacy Check)
// @Tags Mahasiswa
// @Security BearerAuth
// @Param id path string true "ID Mahasiswa"
// @Router /mahasiswa/{id} [get]
func (s *MahasiswaService) GetDetail(c *gin.Context) {
	id := c.Param("id")
	loginUserID := c.GetString("user_id")
	loginRoleID := c.GetString("role_id")

	data, err := s.Repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Mahasiswa tidak ditemukan"})
		return
	}

	isOwner := data.UserID == loginUserID
	isAdvisor := data.AdvisorUserID == loginUserID

	if loginRoleID != RoleAdmin && !isOwner && !isAdvisor {
		c.JSON(403, gin.H{"error": "Forbidden: Anda tidak memiliki akses ke data ini"})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

// GetAchievements godoc
// @Summary Get Mahasiswa Achievements
// @Description Mengambil daftar prestasi mahasiswa
// @Tags Mahasiswa
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID Mahasiswa"
// @Success 200 {object} map[string]interface{}
// @Router /mahasiswa/{id}/achievements [get]
func (s *MahasiswaService) GetAchievements(c *gin.Context) {
	id := c.Param("id")
	loginUserID := c.GetString("user_id")
	loginRoleID := c.GetString("role_id")

	mhs, err := s.Repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Mahasiswa tidak ditemukan"})
		return
	}

	if loginRoleID != RoleAdmin && mhs.UserID != loginUserID && mhs.AdvisorUserID != loginUserID {
		c.JSON(403, gin.H{"error": "Forbidden: Hanya bisa melihat prestasi sendiri atau bimbingan"})
		return
	}

	refs, err := s.AchRepo.FindRefsByStudentID(c.Request.Context(), mhs.UserID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var results []dto.AchievementResponse
	for _, ref := range refs {
		content, _ := s.AchRepo.FindContentByMongoID(c.Request.Context(), ref.MongoAchievementID)
		if content != nil {
			results = append(results, dto.ToAchievementResponse(ref, *content))
		}
	}
	c.JSON(200, gin.H{"data": results})
}

// UpdateAdvisor godoc
// @Summary Update Dosen Wali
// @Description Mengubah dosen wali mahasiswa (Admin Only)
// @Tags Mahasiswa
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID Mahasiswa"
// @Param body body service.UpdateAdvisorRequest true "Request Body"
// @Success 200 {object} map[string]interface{}
// @Router /mahasiswa/{id}/advisor [put]
func (s *MahasiswaService) UpdateAdvisor(c *gin.Context) {
	if c.GetString("role_id") != RoleAdmin {
		c.JSON(403, gin.H{"error": "Forbidden: Hanya Admin"})
		return
	}

	id := c.Param("id")
	var req UpdateAdvisorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}

	if err := s.Repo.UpdateAdvisor(c.Request.Context(), id, req.AdvisorID); err != nil {
		c.JSON(500, gin.H{"error": "Gagal update dosen wali"})
		return
	}
	c.JSON(200, gin.H{"message": "Dosen wali diperbarui"})
}
