package main

import (
	"encoding/json"
	"log"

	"github.com/nats-io/stan.go"
)

func main() {
	sc, err := stan.Connect("test-cluster", "publisher")
	if err != nil {
		log.Fatal(err)
	}
	defer sc.Close()

	order := map[string]interface{}{
		"order_uid":    "b563feb7b2b84b6test",
		"track_number": "WBILMTESTTRACK",
		"entry":        "WBIL",
		"delivery": map[string]interface{}{
			"name":    "Test Testov",
			"phone":   "+9720000000",
			"zip":     "2639809",
			"city":    "Kiryat Mozkin",
			"address": "Ploshad Mira 15",
			"region":  "Kraiot",
			"email":   "test@gmail.com",
		},
		"payment": map[string]interface{}{
			"transaction":   "b563feb7b2b84b6test",
			"request_id":    "",
			"currency":      "USD",
			"provider":      "wbpay",
			"amount":        1817,
			"payment_dt":    1637907727,
			"bank":          "alpha",
			"delivery_cost": 1500,
			"goods_total":   317,
			"custom_fee":    0,
		},
		"items": []map[string]interface{}{
			{
				"chrt_id":      9934930,
				"track_number": "WBILMTESTTRACK",
				"price":        453,
				"rid":          "ab4219087a764ae0btest",
				"name":         "Mascaras",
				"sale":         30,
				"size":         "0",
				"total_price":  317,
				"nm_id":        2389212,
				"brand":        "Vivienne Sabo",
				"status":       202,
			},
		},
		"locale":             "en",
		"internal_signature": "",
		"customer_id":        "test",
		"delivery_service":   "meest",
		"shardkey":           "9",
		"sm_id":              99,
		"date_created":       "2021-11-26T06:22:19Z",
		"oof_shard":          "1",
	}

	data, err := json.Marshal(order)
	if err != nil {
		log.Fatal(err)
	}

	err = sc.Publish("orders", data)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Published message to NATS")
}
