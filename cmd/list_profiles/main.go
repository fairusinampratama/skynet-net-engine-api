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
	
	// Force Direct IP for Randuagung
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

	fmt.Println("Fetching PPP Profiles...")
	res, err := client.Conn.Run("/ppp/profile/print")
	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}

	fmt.Printf("Total Profiles: %d\n", len(res.Re))
	for _, re := range res.Re {
		fmt.Printf("- %s (Local: %s, Remote: %s, Rate: %s)\n", 
			re.Map["name"], re.Map["local-address"], re.Map["remote-address"], re.Map["rate-limit"])
	}
}
