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

	// Check pppoe_users columns
	checkTable(db, "pppoe_users")
	// Check customers columns
	checkTable(db, "customers")
}

func checkTable(db *sql.DB, tableName string) {
	fmt.Printf("--- %s ---\n", tableName)
	rows, err := db.Query("SHOW COLUMNS FROM " + tableName)
	if err != nil {
		fmt.Printf("Table %s not found or error: %v\n", tableName, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var field, typ, null, key, def, extra sql.NullString
		rows.Scan(&field, &typ, &null, &key, &def, &extra)
		fmt.Printf("- %s (%s)\n", field.String, typ.String)
	}
	fmt.Println()
}
