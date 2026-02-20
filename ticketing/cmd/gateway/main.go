package main

import (
	"log"

	"ticketing/internal/common/app"
)

func main() {
	if err := app.Run("gateway"); err != nil {
		log.Fatalf("gateway stopped with error: %v", err)
	}
}
