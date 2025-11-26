package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"pelaporan_prestasi/database"
	"pelaporan_prestasi/route"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading env from environment")
	}

	// init databases
	if err := database.InitPostgres(); err != nil {
		log.Fatal("Postgres init:", err)
	}
	if err := database.InitMongo(); err != nil {
		log.Fatal("Mongo init:", err)
	}

	// start HTTP server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	r := route.SetupRouter()
	log.Printf("Starting server on :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
