package route

import (
	"pelaporan_prestasi/app/service"
	"pelaporan_prestasi/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(r *gin.Engine, authService *service.AuthService, userService *service.UserService, achService *service.AchievementService, mhsService *service.MahasiswaService, dosenService *service.DosenService, reportService *service.ReportService) {

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authService.Login)
			auth.POST("/refresh", authService.Refresh) 
			auth.POST("/logout", authService.Logout)   
			auth.GET("/profile", authService.Profile)  
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
			ach.GET("", achService.GetList)       
			ach.GET("/:id", achService.GetDetail) 
			ach.POST("", achService.Create)      
			ach.PUT("/:id", achService.Update)    
			ach.DELETE("/:id", achService.Delete) 

			ach.POST("/:id/submit", achService.Submit) 
			ach.POST("/:id/verify", achService.Verify) 
			ach.POST("/:id/reject", achService.Reject) 

			ach.GET("/:id/history", achService.GetHistory)            
			ach.POST("/:id/attachments", achService.UploadAttachment) 
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
