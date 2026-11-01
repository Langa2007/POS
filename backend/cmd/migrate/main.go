package main

import (
	"context"
	"log"
	"os"

	"github.com/Langa2007/POS/backend/internal/platform/db"
)

func main() {
	ctx := context.Background()
	store, err := db.Open(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()
	if err := store.Migrate(ctx); err != nil {
		log.Fatal(err)
	}
	log.Print("migrations applied")
}
