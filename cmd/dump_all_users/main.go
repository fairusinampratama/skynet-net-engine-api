package main

import (
	"fmt"
	"log"
	"strings"
	
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

	fmt.Println("Fetching ALL /ppp/active/print...")
	
	res, err := client.Conn.Run("/ppp/active/print")
	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}

	fmt.Printf("Total Active Users: %d\n\n", len(res.Re))
	
	// Group by service type
	pppoe := 0
	ovpn := 0
	l2tp := 0
	other := 0
	
	// Find users with "abdul" or personal names
	fmt.Println("=== Looking for personal users (like abdul) ===")
	for _, re := range res.Re {
		name := re.Map["name"]
		service := re.Map["service"]
		uptime := re.Map["uptime"]
		
		switch service {
		case "pppoe":
			pppoe++
		case "ovpn":
			ovpn++
		case "l2tp":
			l2tp++
		default:
			other++
		}
		
		// Check if name looks like a personal user
		nameLower := strings.ToLower(name)
		if strings.Contains(nameLower, "abdul") || 
		   strings.Contains(nameLower, "ahmad") || 
		   strings.Contains(nameLower, "rt") ||
		   strings.Contains(nameLower, "rw") {
			fmt.Printf("FOUND: %s (service: %s, uptime: %s)\n", name, service, uptime)
		}
	}
	
	fmt.Println("\n=== Service Type Summary ===")
	fmt.Printf("PPPoE: %d\n", pppoe)
	fmt.Printf("OVPN: %d\n", ovpn)
	fmt.Printf("L2TP: %d\n", l2tp)
	fmt.Printf("Other: %d\n", other)
}
