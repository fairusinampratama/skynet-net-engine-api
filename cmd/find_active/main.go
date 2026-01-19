package main

import (
	"fmt"
	
	"skynet-net-engine-api/internal/mikrotik"
	"skynet-net-engine-api/internal/database"
	"skynet-net-engine-api/pkg/logger"
)

func main() {
	logger.Init()
	database.Init()
	
	// 1. Get secrets from Randuagung (Router 1)
	dbUsers, _ := database.GetUsersByRouter(1)
	fmt.Printf("Total secrets for Router 1: %d\n\n", len(dbUsers))
	
	// 2. Get ALL routers
	routers, _ := database.GetAllRouters()
	
	// 3. For each router, check how many Randuagung secrets are active there
	for _, r := range routers {
		// Override for known working IPs
		if r.ID == 1 {
			r.Host = "103.156.128.114"
			r.Port = 8728
		}
		
		client, err := mikrotik.NewClient(r)
		if err != nil {
			fmt.Printf("Router %d (%s): Connection failed - %v\n", r.ID, r.Name, err)
			continue
		}
		
		activeUsers, err := client.GetActiveUsers()
		client.Close()
		
		if err != nil {
			fmt.Printf("Router %d (%s): Query failed\n", r.ID, r.Name)
			continue
		}
		
		// Count how many Randuagung secrets are in this router's active list
		matchCount := 0
		for _, au := range activeUsers {
			if _, exists := dbUsers[au.Name]; exists {
				matchCount++
			}
		}
		
		fmt.Printf("Router %d (%s): %d active, %d match Randuagung secrets\n", 
			r.ID, r.Name, len(activeUsers), matchCount)
	}
}
