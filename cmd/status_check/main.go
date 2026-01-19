package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"skynet-net-engine-api/internal/mikrotik"
	"skynet-net-engine-api/internal/models"
	"skynet-net-engine-api/pkg/logger"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	logger.Init()
	targetIP := "10.2.3.164"
	fmt.Printf("=== CHECKING STATUS FOR %s ===\n", targetIP)

	// 1. Connect to DB
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
	var dbIso bool
	err = db.QueryRow("SELECT username, is_isolated FROM pppoe_users WHERE remote_address = ?", targetIP).Scan(&username, &dbIso)
	if err != nil {
		fmt.Printf("[DB] Error finding user: %v\n", err)
	} else {
		fmt.Printf("[DB] Username: %s | IsIsolated: %v\n", username, dbIso)
	}

	// 2. Connect to Router
	router := models.Router{
		ID:       1,
		Name:     "Randuagung-CCR",
		Host:     "103.156.128.114",
		Port:     8728,
		Username: "skysky",
		Password: "skylineR34!@#",
	}

	client, err := mikrotik.NewClient(router)
	if err != nil {
		log.Fatalf("Failed to connect to router: %v", err)
	}
	defer client.Close()

	// 3. Check Router Firewall
	cmd := []string{"/ip/firewall/address-list/print", "?list=ISOLATED", "?address=" + targetIP}
	res, err := client.Conn.RunArgs(cmd)
	if err != nil {
		log.Fatalf("Failed to check firewall: %v", err)
	}

	if len(res.Re) > 0 {
		fmt.Printf("[ROUTER] User is in ISOLATED Firewall List (Count: %d)\n", len(res.Re))
	} else {
		fmt.Printf("[ROUTER] User is CLEAN (Not in Firewall List)\n")
	}

	// 4. Check Active Session
	cmdActive := []string{"/ppp/active/print", "?address=" + targetIP}
	resActive, err := client.Conn.RunArgs(cmdActive)
	if err != nil {
		log.Fatalf("Failed to check active sessions: %v", err)
	}

	if len(resActive.Re) > 0 {
		fmt.Printf("[ROUTER] User is ONLINE (Uptime: %s)\n", resActive.Re[0].Map["uptime"])
	} else {
		fmt.Printf("[ROUTER] User is OFFLINE\n")
	}
}
