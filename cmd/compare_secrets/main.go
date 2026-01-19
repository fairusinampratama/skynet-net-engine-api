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
	
	// Router 1: Randuagung
	r1 := models.Router{
		ID: 1, 
		Name: "Randuagung", 
		Host: "103.156.128.114", 
		Port: 8728, 
		Username: "skysky", 
		Password: "skylineR34!@#",
	}

	// Router 4: Krian
	r4 := models.Router{
		ID: 4,
		Name: "Krian",
		Host: "103.156.128.114", // Tunnel IP usually, but let's try direct if port mapped?
		Port: 8728, // Wait, Krian is likely behind NAT or VPN
		// We need to use the IP from the database/tunnel 
	}
	// Actually, let's use the DB config to be safe
	// But for this script I'll just check if I can reach Krian logic or just check DB
	// Better yet, just check DB since we synced secrets from both?
	
	// Wait, we only synced Router 1 secrets to DB initially? 
	// Or did we sync all routers?
	// Let's check the database directly.
}
