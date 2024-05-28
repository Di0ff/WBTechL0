package nats

import (
	"database/sql"
	"encoding/json"
	"log"

	"WBTechL0/internal/db"
	stan "github.com/nats-io/stan.go"
)

func SubscribeAndHandle(database *sql.DB, clusterID, clientID, subject string) error {
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL("nats://natsWB:4222"))
	if err != nil {
		return err
	}
	defer sc.Close()

	_, err = sc.Subscribe(subject, func(msg *stan.Msg) {
		var orderData struct {
			Order    db.Order
			Delivery db.Delivery
			Payment  db.Payment
			Items    []db.Item
		}
		log.Println("Received message from NATS:", string(msg.Data))
		err := json.Unmarshal(msg.Data, &orderData)
		if err != nil {
			log.Println("Error unmarshalling message:", err)
			return
		}

		log.Println("Order data received:", orderData)
		err = db.AddOrder(database, orderData.Order, orderData.Delivery, orderData.Payment, orderData.Items)
		if err != nil {
			log.Println("Error adding order to database:", err)
		} else {
			log.Println("Order successfully added to database:", orderData.Order.OrderUID)
		}
	}, stan.DurableName("my-durable"))
	if err != nil {
		return err
	}

	select {}
}
