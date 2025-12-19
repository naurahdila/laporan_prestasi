package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"pelaporan_prestasi/app/repository"
	"pelaporan_prestasi/database"
	"pelaporan_prestasi/middleware"

	"pelaporan_prestasi/route"

	"pelaporan_prestasi/app/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	_ "pelaporan_prestasi/docs" 
)

// @title           Sistem Pelaporan Prestasi API
// @version         1.0
// @description     Dokumentasi API backend pelaporan prestasi mahasiswa.
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"))

	pgPool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("Gagal connect ke PostgreSQL:", err)
	}
	defer pgPool.Close()
	fmt.Println("✅ PostgreSQL Connected!")

	mongoDB, err := database.InitMongo()
	if err != nil {
		log.Fatal("Gagal connect ke MongoDB:", err)
	}
	fmt.Println("✅ MongoDB Connected!")

	userRepo := repository.NewUserRepository(pgPool)
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, os.Getenv("JWT_SECRET"))

	achRepo := repository.NewAchievementRepository(pgPool, mongoDB)
	achService := service.NewAchievementService(achRepo)

	mhsRepo := repository.NewMahasiswaRepository(pgPool)
	mhsService := service.NewMahasiswaService(mhsRepo, achRepo)

	dosenRepo := repository.NewDosenRepository(pgPool)
	dosenService := service.NewDosenService(dosenRepo)

	reportRepo := repository.NewReportRepository(pgPool)
	reportService := service.NewReportService(reportRepo, mhsRepo, achRepo)

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	route.SetupRouter(r, authService, userService, achService, mhsService, dosenService, reportService)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("\n\033[32m=================================================\033[0m")
	log.Println("\033[32m✅  SEMUA DATABASE TERHUBUNG!\033[0m")
	log.Println("\033[32m🚀  SERVER SIAP DI PORT :" + port + "\033[0m")
	log.Println("\033[32m👉  Buka Swagger: http://localhost:" + port + "/swagger/index.html\033[0m")
	log.Println("\033[32m=================================================\033[0m\n")

	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
