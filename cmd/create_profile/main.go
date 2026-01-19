package main

import (
	"fmt"
	"log"
	
	"skynet-net-engine-api/internal/mikrotik"
	"skynet-net-engine-api/internal/models"
	"skynet-net-engine-api/pkg/logger"
)

func main() {
	logger.Init()
	
	router := models.Router{
		ID:       1,
		Name:     "Randuagung-CCR",
		Host:     "103.156.128.114",
		Port:     8728,
		Username: "skysky",
		Password: "skylineR34!@#",
	}

	client, err := mikrotik.NewClient(router)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	fmt.Println("Checking for ISOLATED profile...")
	
	// Check if exists
	res, err := client.Conn.Run("/ppp/profile/print", "?name=ISOLATED")
	if err != nil {
		log.Fatalf("Check failed: %v", err)
	}

	if len(res.Re) > 0 {
		fmt.Println("Profile ISOLATED already exists.")
	} else {
		fmt.Println("Creating ISOLATED profile...")
		// Create with 1k/1k limit (effectively blocked)
		// We could also set a local/remote address from a specific pool if needed, 
		// but rate-limit is the easiest "soft block".
		_, err := client.Conn.Run(
			"/ppp/profile/add", 
			"=name=ISOLATED",
			"=rate-limit=1k/1k",
			"=comment=Created by NetEngine for Isolation",
		)
		if err != nil {
			log.Fatalf("Creation failed: %v", err)
		}
		fmt.Println("Profile ISOLATED created successfully.")
	}
}
