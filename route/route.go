package route

import (
	"pelaporan_prestasi/app/service"
	"pelaporan_prestasi/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(r *gin.Engine, authService *service.AuthService, userService *service.UserService, achService *service.AchievementService, mhsService *service.MahasiswaService, dosenService *service.DosenService, reportService *service.ReportService) {

	// Swagger (Dokumentasi API)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			// Endpoint yang sudah jalan
			auth.POST("/login", authService.Login)

			// Endpoint BARU (Wajib ditambahkan biar gak 404)
			auth.POST("/refresh", authService.Refresh) // Refresh Token
			auth.POST("/logout", authService.Logout)   // Logout
			auth.GET("/profile", authService.Profile)  // Get Profile (Butuh Token)
		}

		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware())

		{
			users.GET("", userService.GetAllUsers)
			users.GET("/:id", userService.GetUserByID)
			users.POST("", userService.CreateUser)
			users.PUT("/:id", userService.UpdateUser)
			users.PUT("/:id/role", userService.UpdateRole)
			users.DELETE("/:id", userService.DeleteUser)
		}

		ach := api.Group("/achievements")
		ach.Use(middleware.AuthMiddleware())
		{
			ach.GET("", achService.GetList)       // 1. List
			ach.GET("/:id", achService.GetDetail) // 2. Detail
			ach.POST("", achService.Create)       // 3. Create
			ach.PUT("/:id", achService.Update)    // 4. Update
			ach.DELETE("/:id", achService.Delete) // 5. Delete

			ach.POST("/:id/submit", achService.Submit) // 6. Submit
			ach.POST("/:id/verify", achService.Verify) // 7. Verify
			ach.POST("/:id/reject", achService.Reject) // 8. Reject

			ach.GET("/:id/history", achService.GetHistory)            // 9. History
			ach.POST("/:id/attachments", achService.UploadAttachment) // 10. Upload
		}

		mahasiswa := api.Group("/mahasiswa")
		mahasiswa.Use(middleware.AuthMiddleware())
		{
			mahasiswa.GET("", mhsService.GetAll)
			mahasiswa.GET("/:id", mhsService.GetDetail)
			mahasiswa.GET("/:id/achievements", mhsService.GetAchievements)
			mahasiswa.PUT("/:id/advisor", mhsService.UpdateAdvisor)
		}

		dosen := api.Group("/dosen")
		dosen.Use(middleware.AuthMiddleware())
		{
			dosen.GET("", dosenService.GetAll)
			dosen.GET("/:id/advisees", dosenService.GetAdvisees)
		}
		
		reports := api.Group("/reports")
        reports.Use(middleware.AuthMiddleware())
        {
            reports.GET("/statistics", reportService.GetGlobalStats)
            reports.GET("/student/:id", reportService.GetStudentReport)
        }
	}
}
