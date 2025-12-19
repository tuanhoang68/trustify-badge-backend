package main

import (
	"log"
	"os"

	"github.com/tuanhoang68/trustify-badge-backend/internal/app"
	"github.com/tuanhoang68/trustify-badge-backend/internal/config"
	"github.com/tuanhoang68/trustify-badge-backend/internal/storage"
)

func main() {
	config.Load()

	db, err := storage.NewDB()
	if err != nil {
		log.Fatal("DB error: ", err)
	}

	r := app.NewRouter(db)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running on :" + port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
