package main

import (
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/db/migrate"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/routes"
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("victory automigration")
	migrate.DoAutoMigrate()
	fmt.Println("victory-frontend.main starting ...")

	fmt.Printf("Listening on: %v -.0 \n", config.Config.Port)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", config.Config.Port), routes.Router()); err != nil {
		log.Fatalf("ListenAndServe failed: %v", err)
	}
}
