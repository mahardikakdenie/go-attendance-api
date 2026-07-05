package main

import (
	"fmt"
	"log"
	"os"

	"go-attendance-api/internal/config"
	"go-attendance-api/internal/seeder"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: File .env tidak ditemukan")
	}

	os.Setenv("RUN_SEEDER", "true")
	db := config.InitDB()

	fmt.Println("Running SeedRoles explicitly...")
	seeder.SeedRoles(db)
	fmt.Println("SeedRoles finished successfully!")
}
