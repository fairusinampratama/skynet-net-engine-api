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

	// 2. Connect to Router
	router := models.Router{
		ID:       1,
		Name:     "Randuagung-CCR",
		Host:     "103.156.128.114",
		Port:     8728,
		Username: "skysky",
		Password: "skylineR34!@#",
	}

	fmt.Printf("Dialing %s (%s)...\n", router.Name, router.Host)
	client, err := mikrotik.NewClient(router)
	if err != nil {
		log.Fatalf("Failed to connect to router: %v", err)
	}
	defer client.Close()

	targetIP := "10.2.3.164"

	// 3. Check Router Firewall
	fmt.Println("Checking Firewall Address List...")
	cmd := []string{"/ip/firewall/address-list/print", "?list=ISOLATED", "?address=" + targetIP, "=.proplist=.id"}
	res, err := client.Conn.RunArgs(cmd)
	if err != nil {
		log.Fatalf("Failed check firewall: %v", err)
	}

	routerIsolated := len(res.Re) > 0
	fmt.Printf("Router Status: Isolated=%v\n", routerIsolated)

	if routerIsolated {
		fmt.Println("User is Isolated on Router. Removing...")
		for _, re := range res.Re {
			id := re.Map[".id"]
			_, err := client.Conn.Run("/ip/firewall/address-list/remove", "=.id="+id)
			if err != nil {
				log.Printf("Failed to remove ID %s: %v", id, err)
			} else {
				fmt.Printf("Removed ID %s\n", id)
			}
		}
	} else {
		fmt.Println("User is NOT Isolated on Router.")
	}

	// 4. Update DB to match Reality (Not Isolated)
	fmt.Println("Syncing Database...")
	_, err = db.Exec("UPDATE pppoe_users SET is_isolated = FALSE WHERE remote_address = ?", targetIP)
	if err != nil {
		log.Fatalf("Failed to update DB: %v", err)
	}
	fmt.Println("Database Updated: is_isolated = FALSE")
}
