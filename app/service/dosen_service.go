package service

import (
	"pelaporan_prestasi/app/repository"
	"github.com/gin-gonic/gin"
)

type DosenService struct {
	Repo *repository.DosenRepository
}

func NewDosenService(repo *repository.DosenRepository) *DosenService {
	return &DosenService{Repo: repo}
}

// GetAll godoc
// @Summary Get All Dosen
// @Tags Dosen
// @Security BearerAuth
// @Router /dosen [get]
func (s *DosenService) GetAll(c *gin.Context) {
    loginUserID := c.GetString("user_id")
    loginRoleID := c.GetString("role_id")

    data, err := s.Repo.GetAll(c.Request.Context())
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    var filteredData []repository.DosenData
    for _, d := range data {
        if loginRoleID == RoleAdmin {
            filteredData = append(filteredData, d)
        } else if loginRoleID == RoleDosen && d.UserID == loginUserID {
            filteredData = append(filteredData, d)
        }
    }

    if filteredData == nil {
        filteredData = []repository.DosenData{}
    }
    c.JSON(200, gin.H{"data": filteredData})
}

// GetAdvisees godoc
// @Summary Get Dosen Advisees (With Authorization)
// @Tags Dosen
// @Security BearerAuth
// @Param id path string true "ID Dosen (Tabel Dosen)"
// @Router /dosen/{id}/advisees [get]
func (s *DosenService) GetAdvisees(c *gin.Context) {
	dosenID := c.Param("id") 
	loginUserID := c.GetString("user_id")
	loginRoleID := c.GetString("role_id")
	dosen, err := s.Repo.GetByID(c.Request.Context(), dosenID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Dosen tidak ditemukan"})
		return
	}

	if loginRoleID == RoleDosen && dosen.UserID != loginUserID {
		c.JSON(403, gin.H{"error": "Forbidden: Anda tidak diperbolehkan melihat bimbingan dosen lain"})
		return
	}
	
	data, err := s.Repo.GetAdvisees(c.Request.Context(), dosenID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": data})
}