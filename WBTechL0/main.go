package main

import (
	"database/sql"
	"log"

	"WBTechL0/internal/cache"
	"WBTechL0/internal/http"
	"WBTechL0/internal/nats"
	_ "github.com/lib/pq"
)

func main() {
	dbConn, err := sql.Open("postgres", "postgres://admin:1111@postgres:5432/ordersWB?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	orderCache := cache.NewCache(dbConn)

	err = orderCache.LoadCacheFromDB()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err := nats.SubscribeAndHandle(dbConn, "test-cluster", "client-id", "orders")
		if err != nil {
			log.Fatal(err)
		}
	}()

	http.StartServer(orderCache)
}
