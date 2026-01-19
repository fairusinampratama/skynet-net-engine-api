package main

import (
	"fmt"
	"log"
	
	"skynet-net-engine-api/internal/mikrotik"
	"skynet-net-engine-api/internal/database"
	"skynet-net-engine-api/pkg/logger"
)

func main() {
	logger.Init()
	
	// 1. Get All Routers
	database.Init() // Uses default credentials from env or fallback
	routers, err := database.GetAllRouters()
	if err != nil {
		log.Fatalf("Failed to fetch routers: %v", err)
	}
	
	targetIP := "10.2.3.164"
	fmt.Printf("Scanning %d routers for IP: %s\n", len(routers), targetIP)
	
	for _, r := range routers {
		// Quick timeout for scanning
		// Note: We might need to handle the tunnel host mapping if not running inside the tunnel?
		// For now, try default. If locally running, tunnels might be accessible.
		
		fmt.Printf("Checking %s (%s)...\n", r.Name, r.Host)
		if r.ID == 1 {
			r.Host = "103.156.128.114" // Override for Randuagung as known good
			r.Port = 8728
		}

		client, err := mikrotik.NewClient(r)
		if err != nil {
			fmt.Printf(" - Connection Failed: %v\n", err)
			continue
		}
		
		// Check Active
		res, err := client.Conn.Run("/ppp/active/print", "?address="+targetIP)
		if err != nil {
			fmt.Printf(" - Query Failed: %v\n", err)
			client.Close()
			continue
		}
		
		if len(res.Re) > 0 {
			user := res.Re[0]
			fmt.Printf("!!! FOUND USER ON ROUTER %d (%s) !!!\n", r.ID, r.Name)
			fmt.Printf(" - Name: %s\n", user.Map["name"])
			fmt.Printf(" - Uptime: %s\n", user.Map["uptime"])
			client.Close()
			return
		}
		client.Close()
	}
	
	fmt.Println("Scan Complete. User not found on ANY accessible router.")
}
