package main

import (
	"log"
	"fmt"
	"database/sql" // Added import
	"skynet-net-engine-api/internal/database"
	"skynet-net-engine-api/pkg/logger" // Added import
	
	_ "github.com/go-sql-driver/mysql" // Added driver
)

func main() {
	// Initialize Logger first
	logger.Init()
	
	// Open DB manually
	dsn := "fairusinampratama@tcp(127.0.0.1:3306)/netengine?parseTime=true"
	var err error
	database.DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	
	routers, err := database.GetAllRouters()
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Found %d routers\n", len(routers))
	found := false
	for _, r := range routers {
		fmt.Printf("ID: %d, Name: %s, Host: %s, Port: %d\n", r.ID, r.Name, r.Host, r.Port)
		if r.ID == 1 {
			found = true
		}
	}
	
	if found {
		fmt.Println("SUCCESS: Router 1 is in the list with correct config.")
	} else {
		fmt.Println("FAILURE: Router 1 is MISSING from the list.")
	}
}
