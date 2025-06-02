package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/vit0-9/utils_api/pkg/utils"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("WARN: Error loading .env file, using environment variables from system if set.")
	}
	cityDBPath := os.Getenv("MMDB_CITY_PATH")
	asnDBPath := os.Getenv("MMDB_ASN_PATH")

	utils.LoadMaxMindDBs(cityDBPath, asnDBPath)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")
		utils.CloseMaxMindDBs() // Close both databases
		os.Exit(0)
	}()

	app, err := NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	if err := app.Start(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		utils.CloseMaxMindDBs()
	}
}
