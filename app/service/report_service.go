package service

import (
    "pelaporan_prestasi/app/models/dto" // Pastikan import DTO untuk mapping response
    "pelaporan_prestasi/app/repository"
    "github.com/gin-gonic/gin"
)

type ReportService struct {
    Repo    *repository.ReportRepository
    MhsRepo *repository.MahasiswaRepository
    AchRepo *repository.AchievementRepository // Tambahkan ini untuk akses data prestasi
}

func NewReportService(repo *repository.ReportRepository, mhsRepo *repository.MahasiswaRepository, achRepo *repository.AchievementRepository) *ReportService {
    return &ReportService{Repo: repo, MhsRepo: mhsRepo, AchRepo: achRepo}
}

// GetGlobalStats godoc
// @Summary Get Global Statistics
// @Tags Reports
// @Security BearerAuth
// @Router /api/v1/reports/statistics [get]
func (s *ReportService) GetGlobalStats(c *gin.Context) {
    loginUserID := c.GetString("user_id")
    loginRoleID := c.GetString("role_id")

    if loginRoleID == RoleAdmin {
        stats, err := s.Repo.GetGlobalStats(c.Request.Context())
        if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
        c.JSON(200, gin.H{"scope": "Global", "data": stats})

    } else if loginRoleID == RoleMahasiswa {
        stats, err := s.Repo.GetStudentStats(c.Request.Context(), loginUserID)
        if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
        c.JSON(200, gin.H{"scope": "Personal", "data": stats})

    } else if loginRoleID == RoleDosen {
        // Logika Dosen Wali bisa ditambahkan di sini untuk melihat statistik bimbingan
        c.JSON(200, gin.H{"scope": "Dosen Wali", "message": "Statistik bimbingan Anda"})
    }
}

// GetStudentReport godoc
// @Summary Get Individual Student Report
// @Tags Reports
// @Security BearerAuth
// @Param id path string true "ID Mahasiswa"
// @Router /api/v1/reports/student/{id} [get]
func (s *ReportService) GetStudentReport(c *gin.Context) {
    studentID := c.Param("id")
    loginUserID := c.GetString("user_id")
    loginRoleID := c.GetString("role_id")

    // 1. Ambil info profil mahasiswa
    mhs, err := s.MhsRepo.GetByID(c.Request.Context(), studentID)
    if err != nil {
        c.JSON(404, gin.H{"error": "Mahasiswa tidak ditemukan"})
        return
    }

    // 2. Cek Privasi
    if loginRoleID == RoleMahasiswa && mhs.UserID != loginUserID {
        c.JSON(403, gin.H{"error": "Forbidden: Anda hanya bisa melihat laporan Anda sendiri"})
        return
    }

    // 3. Ambil data prestasi LANGSUNG
    refs, _ := s.AchRepo.FindRefsByStudentID(c.Request.Context(), mhs.UserID)
    var achievements []dto.AchievementResponse
    
    for _, ref := range refs {
        content, _ := s.AchRepo.FindContentByMongoID(c.Request.Context(), ref.MongoAchievementID)
        if content != nil {
            achievements = append(achievements, dto.ToAchievementResponse(ref, *content))
        }
    }

    // 4. Tampilkan semua data dalam satu response
    c.JSON(200, gin.H{
        "student_info": mhs,
        "achievements": achievements, // Data prestasi muncul di sini, bukan cuma link
    })
}