package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"skynet-net-engine-api/pkg/logger"
)

func main() {
	logger.Init()
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "fairusinampratama@tcp(127.0.0.1:3306)/netengine?parseTime=true"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	fmt.Println("Syncing is_isolated based on profile name...")
	
	// Set is_isolated = TRUE where profile = 'isolirebilling'
	result, err := db.Exec("UPDATE pppoe_users SET is_isolated = (profile = 'isolirebilling')")
	if err != nil {
		log.Fatalf("Sync failed: %v", err)
	}
	
	rows, _ := result.RowsAffected()
	fmt.Printf("Updated %d rows.\n", rows)
	
	// Count how many are now isolated
	var count int
	db.QueryRow("SELECT COUNT(*) FROM pppoe_users WHERE is_isolated = TRUE").Scan(&count)
	fmt.Printf("Total isolated users: %d\n", count)
}
