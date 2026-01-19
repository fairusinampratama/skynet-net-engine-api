package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "fairusinampratama@tcp(127.0.0.1:3306)/netengine?parseTime=true"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	var username string
	var ip string
	var isIsolated bool

	err = db.QueryRow("SELECT username, remote_address, is_isolated FROM pppoe_users WHERE remote_address = ?", "10.2.3.164").Scan(&username, &ip, &isIsolated)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	fmt.Printf("User: %s | IP: %s | IsIsolated: %v\n", username, ip, isIsolated)
}
