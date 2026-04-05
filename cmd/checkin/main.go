package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/stockyard-dev/stockyard-checkin/internal/server"
	"github.com/stockyard-dev/stockyard-checkin/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9807"
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./checkin-data"
	}

	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("checkin: %v", err)
	}
	defer db.Close()

	srv := server.New(db, server.DefaultLimits())

	fmt.Printf("\n  Checkin — Self-hosted member check-in and attendance tracking\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n  Questions? hello@stockyard.dev — I read every message\n\n", port, port)
	log.Printf("checkin: listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, srv))
}
