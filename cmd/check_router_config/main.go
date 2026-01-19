package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "fairusinampratama@tcp(127.0.0.1:3306)/netengine?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Fetch Randuagung-CCR config
	var id int
	var name, host, username, password string
	var port int
	
	// Assuming ID 1 based on previous logs, but let's search by name too if needed
	query := "SELECT id, name, host, port, username, password FROM routers WHERE id = 1"
	
	err = db.QueryRow(query).Scan(&id, &name, &host, &port, &username, &password)
	if err != nil {
		log.Fatal("Failed to fetch router: ", err)
	}

	fmt.Printf("Router Config:\nID: %d\nName: %s\nHost: %s\nPort: %d\nUser: %s\nPass: %s\n", 
		id, name, host, port, username, password)
}
