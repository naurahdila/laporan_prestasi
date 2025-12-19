package service

import (
	"fmt"
	"os"
	"path/filepath"
	"pelaporan_prestasi/app/models/dto"
	mongodb "pelaporan_prestasi/app/models/mongo"
	"pelaporan_prestasi/app/models/postgres"
	"pelaporan_prestasi/app/repository"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AchievementService struct {
	Repo *repository.AchievementRepository
}

func NewAchievementService(repo *repository.AchievementRepository) *AchievementService {
	return &AchievementService{Repo: repo}
}

// Request Body Structs
type CreateAchievementRequest struct {
	Title           string   `form:"title" binding:"required"`
	Description     string   `form:"description" binding:"required"`
	AchievementType string   `form:"achievement_type" binding:"required"`
	Points          int      `form:"points"`
	Tags            []string `form:"tags"`
}

type VerifyRequest struct {
	Status string `json:"status" binding:"required" example:"VERIFIED"`
	Notes  string `json:"notes" example:"Oke bagus"`
}

// --- 1. LIST ---
// GetList godoc
// @Summary List Achievements
// @Tags Achievements
// @Security BearerAuth
// @Success 200 {array} dto.AchievementResponse
// @Router /achievements [get]
func (s *AchievementService) GetList(c *gin.Context) {
	roleID := c.GetString("role_id")
	userID := c.GetString("user_id")

	var refs []postgres.AchievementReference
	var err error

	if roleID == RoleMahasiswa {
		refs, err = s.Repo.FindRefsByStudentID(c.Request.Context(), userID)
	} else if roleID == RoleDosen {
		refs, err = s.Repo.FindRefsByAdvisorID(c.Request.Context(), userID)
	} else if roleID == RoleAdmin {
		refs, err = s.Repo.FindAllRefs(c.Request.Context())
	} else {
		c.JSON(403, gin.H{"error": "Role invalid"})
		return
	}

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var results []dto.AchievementResponse
	for _, ref := range refs {
		content, _ := s.Repo.FindContentByMongoID(c.Request.Context(), ref.MongoAchievementID)
		if content != nil {
			results = append(results, dto.ToAchievementResponse(ref, *content))
		}
	}
	c.JSON(200, gin.H{"data": results})
}

// --- 2. DETAIL ---
// GetDetail godoc
// @Summary Get Detail
// @Tags Achievements
// @Security BearerAuth
// @Param id path string true "ID"
// @Success 200 {object} dto.AchievementResponse
// @Router /achievements/{id} [get]
func (s *AchievementService) GetDetail(c *gin.Context) {
	id := c.Param("id")
	ref, err := s.Repo.FindRefByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	if c.GetString("role_id") == RoleMahasiswa && ref.StudentID != c.GetString("user_id") {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	content, err := s.Repo.FindContentByMongoID(c.Request.Context(), ref.MongoAchievementID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Content missing"})
		return
	}

	c.JSON(200, gin.H{"data": dto.ToAchievementResponse(*ref, *content)})
}

// --- 3. CREATE ---
// Create godoc
// @Summary      Create Achievement with File
// @Description  Kirim data + file sekaligus (Multipart Form)
// @Tags         Achievements
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Param        title formData string true "Judul"
// @Param        description formData string true "Deskripsi"
// @Param        achievement_type formData string true "Tipe"
// @Param        points formData int false "Poin"
// @Param        file formData file false "File Bukti (PDF/Image)"
// @Success      201 {object} map[string]interface{}
// @Router       /achievements [post]
func (s *AchievementService) Create(c *gin.Context) {
	if c.GetString("role_id") != RoleMahasiswa {
		c.JSON(403, gin.H{"error": "Hanya mahasiswa yang boleh upload"})
		return
	}

	// 1. Bind Form Data (Bukan JSON lagi)
	var req CreateAchievementRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": "Input Salah: " + err.Error()})
		return
	}

	// 2. Handle File Upload (Optional/Required terserah)
	var attachments []mongodb.Attachment
	file, errFile := c.FormFile("file") // Ambil file dengan key "file"

	if errFile == nil { // Jika ada file yang diupload
		// Validasi Ekstensi
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext != ".pdf" && ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			c.JSON(400, gin.H{"error": "Format file harus PDF atau Gambar!"})
			return
		}

		// Buat Folder Upload jika belum ada
		uploadDir := "uploads"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.Mkdir(uploadDir, 0755)
		}

		// Simpan File
		filename := fmt.Sprintf("%s_%s%s", c.GetString("user_id"), uuid.New().String(), ext)
		savePath := filepath.Join(uploadDir, filename)
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(500, gin.H{"error": "Gagal simpan file"})
			return
		}

		// Tambahkan ke array attachments
		attachments = append(attachments, mongodb.Attachment{
			FileName:   file.Filename,
			FileURL:    savePath,
			FileType:   ext,
			UploadedAt: time.Now(),
		})
	}

	// 3. Simpan ke MongoDB
	mongoData := mongodb.Achievement{
		StudentID:       c.GetString("user_id"),
		Title:           req.Title,
		Description:     req.Description,
		AchievementType: req.AchievementType,
		Tags:            req.Tags,
		Points:          req.Points,
		Attachments:     attachments,                  // Masukkan file tadi
		Details:         make(map[string]interface{}), // Kosong dulu biar aman
	}

	mongoID, err := s.Repo.InsertMongo(c.Request.Context(), &mongoData)
	if err != nil {
		c.JSON(500, gin.H{"error": "Mongo Error: " + err.Error()})
		return
	}

	// 4. Simpan ke Postgres
	pgRef := postgres.AchievementReference{
		StudentID:          c.GetString("user_id"),
		MongoAchievementID: mongoID,
	}
	// Perbaikan: Cek error Postgres dengan teliti
	if err := s.Repo.InsertPostgres(c.Request.Context(), &pgRef); err != nil {
		// Jika ini error, berarti FK Constraint (User tidak ada) atau DB Error
		c.JSON(500, gin.H{"error": "Postgres Reference Error: " + err.Error()})
		return
	}

	// 5. Simpan History (Perbaikan: Cek Error!)
	hist := postgres.AchievementHistory{
		AchievementID:  pgRef.ID,
		ChangedBy:      c.GetString("user_id"),
		PreviousStatus: "NONE",
		NewStatus:      "DRAFT",
		Remarks:        "Created with file",
	}
	if err := s.Repo.AddHistory(c.Request.Context(), hist); err != nil {
		// Kita log errornya, tapi jangan gagalkan request utama (opsional)
		fmt.Printf("⚠️ Gagal catat history: %v\n", err)
		// Kalau mau strict: c.JSON(500, ...); return
	}

	c.JSON(201, gin.H{
		"status":  "success",
		"message": "Prestasi berhasil dibuat + File terupload",
		"id":      pgRef.ID,
	})
}

// --- 4. UPDATE ---
// Update godoc
// @Summary Update Achievement Content
// @Tags Achievements
// @Security BearerAuth
// @Param id path string true "ID"
// @Param body body CreateAchievementRequest true "Body"
// @Router /achievements/{id} [put]
func (s *AchievementService) Update(c *gin.Context) {
	id := c.Param("id")
	ref, err := s.Repo.FindRefByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	if c.GetString("role_id") == RoleMahasiswa {
		if ref.StudentID != c.GetString("user_id") {
			c.JSON(403, gin.H{"error": "Forbidden"})
			return
		}
		if ref.Status != "DRAFT" && ref.Status != "REJECTED" {
			c.JSON(400, gin.H{"error": "Locked"})
			return
		}
	}

	// Kembali pakai JSON untuk update teks saja (atau mau ubah jadi form juga boleh)
	// Di sini saya biarkan JSON agar konsisten dengan request sebelumnya
	// Kecuali Mas mau update file juga di sini, maka harus ganti binding form.
	type UpdateReq struct {
		Title           string   `json:"title"`
		Description     string   `json:"description"`
		AchievementType string   `json:"achievement_type"`
		Points          int      `json:"points"`
		Tags            []string `json:"tags"`
	}

	var req UpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updateData := mongodb.Achievement{
		Title:           req.Title,
		Description:     req.Description,
		AchievementType: req.AchievementType,
		Tags:            req.Tags,
		Points:          req.Points,
	}
	s.Repo.UpdateContentMongo(c.Request.Context(), ref.MongoAchievementID, updateData)
	c.JSON(200, gin.H{"message": "Updated"})
}

// --- 5. DELETE ---
// Delete godoc
// @Summary Delete Achievement
// @Tags Achievements
// @Security BearerAuth
// @Param id path string true "ID"
// @Router /achievements/{id} [delete]
func (s *AchievementService) Delete(c *gin.Context) {
	id := c.Param("id")
	ref, err := s.Repo.FindRefByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	if ref.StudentID != c.GetString("user_id") {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}
	if ref.Status != "DRAFT" {
		c.JSON(400, gin.H{"error": "Locked"})
		return
	}

	s.Repo.DeleteRef(c.Request.Context(), id)
	c.JSON(200, gin.H{"message": "Deleted"})
}

// --- 6. SUBMIT ---
// Submit godoc
// @Summary Submit for Verification
// @Tags Achievements
// @Security BearerAuth
// @Param id path string true "ID"
// @Router /achievements/{id}/submit [post]
func (s *AchievementService) Submit(c *gin.Context) {
	id := c.Param("id")
	ref, err := s.Repo.FindRefByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}
	if ref.StudentID != c.GetString("user_id") {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	s.Repo.UpdateStatus(c.Request.Context(), id, "PENDING", nil)
	s.Repo.AddHistory(c.Request.Context(), postgres.AchievementHistory{AchievementID: id, ChangedBy: c.GetString("user_id"), PreviousStatus: ref.Status, NewStatus: "PENDING", Remarks: "Submitted"})
	c.JSON(200, gin.H{"message": "Submitted"})
}

// --- 7. VERIFY ---
// Verify godoc
// @Summary Verify Achievement (Dosen)
// @Tags Achievements
// @Security BearerAuth
// @Param id path string true "ID"
// @Param body body VerifyRequest true "Body"
// @Router /achievements/{id}/verify [post]
func (s *AchievementService) Verify(c *gin.Context) {
	id := c.Param("id")
	dosenID := c.GetString("user_id")

	isAdvisor, _ := s.Repo.IsAdvisorOfRef(c.Request.Context(), id, dosenID)
	if c.GetString("role_id") == RoleAdmin {
		isAdvisor = true
	}
	if !isAdvisor {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ref, _ := s.Repo.FindRefByID(c.Request.Context(), id)
	if req.Status == "VERIFIED" {
		s.Repo.UpdateVerified(c.Request.Context(), id, dosenID)
	} else {
		s.Repo.UpdateStatus(c.Request.Context(), id, "REJECTED", &req.Notes)
	}
	s.Repo.AddHistory(c.Request.Context(), postgres.AchievementHistory{AchievementID: id, ChangedBy: dosenID, PreviousStatus: ref.Status, NewStatus: req.Status, Remarks: req.Notes})
	c.JSON(200, gin.H{"message": "Status updated"})
}

// --- 8. REJECT ---
// Reject godoc
// @Summary Reject Achievement (Dosen)
// @Tags Achievements
// @Security BearerAuth
// @Param id path string true "ID"
// @Param body body VerifyRequest true "Body"
// @Router /achievements/{id}/reject [post]
func (s *AchievementService) Reject(c *gin.Context) {
	// Reusing Verify Logic but explicit endpoint
	s.Verify(c)
}

// --- 9. HISTORY ---
// GetHistory godoc
// @Summary Get History
// @Tags Achievements
// @Security BearerAuth
// @Param id path string true "ID"
// @Router /achievements/{id}/history [get]
func (s *AchievementService) GetHistory(c *gin.Context) {
	id := c.Param("id")
	hist, _ := s.Repo.GetHistory(c.Request.Context(), id)
	c.JSON(200, gin.H{"data": hist})
}

// --- 10. UPLOAD ---
// UploadAttachment godoc
// @Summary Upload Attachment
// @Description Upload file bukti (bisa berkali-kali).
// @Tags Achievements
// @Security BearerAuth
// @Accept multipart/form-data
// @Param id path string true "ID"
// @Param file formData file true "File"
// @Router /achievements/{id}/attachments [post]
func (s *AchievementService) UploadAttachment(c *gin.Context) {
	id := c.Param("id")
	ref, err := s.Repo.FindRefByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}
	if c.GetString("role_id") == RoleMahasiswa && ref.StudentID != c.GetString("user_id") {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "File missing"})
		return
	}

	filename := fmt.Sprintf("%s_%s%s", id, uuid.New().String(), filepath.Ext(file.Filename))
	savePath := filepath.Join("uploads", filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(500, gin.H{"error": "Save failed"})
		return
	}

	newAttachment := mongodb.Attachment{
		FileName:   file.Filename,
		FileURL:    savePath,
		FileType:   filepath.Ext(file.Filename),
		UploadedAt: time.Now(),
	}

	if err := s.Repo.AddAttachmentMongo(c.Request.Context(), ref.MongoAchievementID, newAttachment); err != nil {
		c.JSON(500, gin.H{"error": "DB Update failed"})
		return
	}

	c.JSON(200, gin.H{"message": "File attached", "data": newAttachment})
}
