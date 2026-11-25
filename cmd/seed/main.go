package main

import (
	"context"
	"log"
	"os"

	"typetype/internal/database"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	DATABASE_URL := "postgresql://postgres:1411621570@localhost:5432/typetype"
	// Database connection
	pool, err := pgxpool.New(ctx, DATABASE_URL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	// Initialize SQLC Queries
	db := database.New(pool)

	// Read file
	content, err := os.ReadFile(
		"C:/Users/pc/Codes/typetype-frontend/public/dena_pawna.txt",
	)
	if err != nil {
		log.Fatal("Failed to read file:", err)
	}

	// Insert passage
	_, err = db.CreatePassage(ctx, database.CreatePassageParams{
		Content:  string(content),
		Language: "bn",
	})
	if err != nil {
		log.Fatal("Failed to create passage:", err)
	}

	log.Println("Passage imported successfully!")
}
