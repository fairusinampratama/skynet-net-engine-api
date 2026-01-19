package main

import (
	"fmt"
	"log"
	"skynet-net-engine-api/internal/mikrotik"
	"skynet-net-engine-api/internal/models"
	"skynet-net-engine-api/pkg/logger"
)

func main() {
	logger.Init() // Correct function name
	
	// Force Direct IP for Randuagung
	router := models.Router{
		ID:       1,
		Name:     "Randuagung-CCR",
		Host:     "103.156.128.114",
		Port:     8728,
		Username: "skysky",
		Password: "skylineR34!@#",
	}

	fmt.Printf("Dialing %s (%s)...\n", router.Name, router.Host)
	client, err := mikrotik.NewClient(router)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	fmt.Println("Connected! Fetching Firewall Address List...")
	// We need a method to get address list. The client might not have it exposed directly as 'GetAddressList' 
	// based on previous file views (I only saw Add/Remove).
	// usage: /ip/firewall/address-list/print where list=ISOLATED
	
	cmd := []string{"/ip/firewall/address-list/print", "?list=ISOLATED"}
	res, err := client.Client.RunArgs(cmd) // Access inner Client field
	if err != nil {
		log.Fatalf("Failed to fetch address list: %v", err)
	}

	fmt.Printf("Total Isolated Entries: %d\n", len(res.Re))
	found := false
	for _, re := range res.Re {
		address := re.Map["address"]
		if address == "10.2.3.164" {
			fmt.Printf("FOUND IN FIREWALL: %s | List: %s\n", address, re.Map["list"])
			found = true
		}
	}

	if !found {
		fmt.Println("User 10.2.3.164 NOT FOUND in ISOLATED firewall list.")
	}
}
