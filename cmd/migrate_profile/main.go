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

	fmt.Println("Checking for previous_profile column...")
	query := "SHOW COLUMNS FROM pppoe_users LIKE 'previous_profile'"
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Check failed: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		fmt.Println("Adding previous_profile column...")
		_, err := db.Exec("ALTER TABLE pppoe_users ADD COLUMN previous_profile VARCHAR(64) DEFAULT NULL")
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("Column previous_profile added successfully.")
	} else {
		fmt.Println("Column previous_profile already exists.")
	}
}
