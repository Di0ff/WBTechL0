package nats

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"WBTechL0/internal/db"
	"github.com/nats-io/stan.go"
)

type OrderData struct {
	OrderUID          string      `json:"order_uid"`
	TrackNumber       string      `json:"track_number"`
	Entry             string      `json:"entry"`
	Locale            string      `json:"locale"`
	InternalSignature string      `json:"internal_signature"`
	CustomerID        string      `json:"customer_id"`
	DeliveryService   string      `json:"delivery_service"`
	ShardKey          string      `json:"shardkey"`
	SMID              int         `json:"sm_id"`
	DateCreated       string      `json:"date_created"`
	OOFShard          string      `json:"oof_shard"`
	Delivery          db.Delivery `json:"delivery"`
	Payment           db.Payment  `json:"payment"`
	Items             []db.Item   `json:"items"`
}

func SubscribeAndHandle(database *sql.DB, clusterID, clientID, subject string) error {
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL("nats://natsWB:4222"))
	if err != nil {
		log.Println("Error connecting to NATS Streaming server:", err)
		return err
	}
	defer sc.Close()

	_, err = sc.Subscribe(subject, func(msg *stan.Msg) {
		log.Println("Received message from NATS:", string(msg.Data))

		var orderData OrderData

		err := json.Unmarshal(msg.Data, &orderData)
		if err != nil {
			log.Println("Error unmarshalling message:", err)
			return
		}

		log.Printf("Order data after unmarshalling: %+v\n", orderData)

		dateCreated, err := time.Parse(time.RFC3339, orderData.DateCreated)
		if err != nil {
			log.Println("Error parsing date:", err)
			return
		}

		order := db.Order{
			OrderUID:          orderData.OrderUID,
			TrackNumber:       orderData.TrackNumber,
			Entry:             orderData.Entry,
			Locale:            orderData.Locale,
			InternalSignature: orderData.InternalSignature,
			CustomerID:        orderData.CustomerID,
			DeliveryService:   orderData.DeliveryService,
			ShardKey:          orderData.ShardKey,
			SMID:              orderData.SMID,
			DateCreated:       dateCreated,
			OOFShard:          orderData.OOFShard,
		}

		delivery := db.Delivery{
			OrderUID: orderData.OrderUID,
			Name:     orderData.Delivery.Name,
			Phone:    orderData.Delivery.Phone,
			Zip:      orderData.Delivery.Zip,
			City:     orderData.Delivery.City,
			Address:  orderData.Delivery.Address,
			Region:   orderData.Delivery.Region,
			Email:    orderData.Delivery.Email,
		}

		log.Printf("OrderUID for delivery before insert: %s\n", delivery.OrderUID)

		err = db.AddOrder(database, order, delivery, orderData.Payment, orderData.Items)
		if err != nil {
			log.Println("Error adding order to database:", err)
		} else {
			log.Println("Order successfully added to database:", order.OrderUID)
		}
	}, stan.DurableName("my-durable"))
	if err != nil {
		log.Println("Error subscribing to subject:", err)
		return err
	}

	log.Println("Subscribed to NATS subject:", subject)

	select {}
}
