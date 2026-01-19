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

	fmt.Println("Fetching RAW /ppp/active/print (First 3 entries)...")
	
	// Access inner client to run raw command without proplist
	res, err := client.Conn.Run("/ppp/active/print")
	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}

	fmt.Printf("Total Entries: %d\n", len(res.Re))
	for i, re := range res.Re {
		if i >= 3 { break }
		fmt.Printf("--- Entry %d ---\n", i+1)
		for k, v := range re.Map {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}
}
