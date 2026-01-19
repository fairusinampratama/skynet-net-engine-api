package main

import (
	"fmt"
	"skynet-net-engine-api/internal/database"
	"skynet-net-engine-api/pkg/logger"
    "log"
)

func main() {
    logger.Init()
	database.Init()
	routers, err := database.GetAllRouters()
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range routers {
		fmt.Printf("ID: %d | Name: %s | Host: %s\n", r.ID, r.Name, r.Host)
	}
}
