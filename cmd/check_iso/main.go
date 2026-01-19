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

	var isIsolated bool
	query := "SELECT is_isolated FROM pppoe_users WHERE remote_address = '10.2.3.164'"
	err = db.QueryRow(query).Scan(&isIsolated)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("User not found or no isolation status set.")
		} else {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("Isolation Status for 10.2.3.164: %v\n", isIsolated)
	}
}
