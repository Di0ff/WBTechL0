package http

import (
	"WBTechL0/internal/cache"
	"WBTechL0/internal/db"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func StartServer(orderCache *cache.Cache) {
	http.Handle("/", http.FileServer(http.Dir("./assets")))

	http.HandleFunc("/order/", func(w http.ResponseWriter, r *http.Request) {
		orderID := r.URL.Path[len("/order/"):]
		log.Println("Received request for order:", orderID)
		if orderID == "" {
			http.Error(w, "Order ID is required", http.StatusBadRequest)
			return
		}

		order, err := orderCache.GetOrder(orderID)
		if err != nil {
			http.Error(w, "Order not found", http.StatusNotFound)
			log.Println("Order not found in cache or DB:", orderID)
			return
		}

		delivery, err := orderCache.GetDelivery(orderID)
		if err != nil {
			http.Error(w, "Delivery information not found", http.StatusNotFound)
			log.Println("Delivery information not found for order:", orderID)
			return
		}

		payment, err := orderCache.GetPayment(orderID)
		if err != nil {
			http.Error(w, "Payment information not found", http.StatusNotFound)
			log.Println("Payment information not found for order:", orderID)
			return
		}

		items, err := orderCache.GetItems(orderID)
		if err != nil {
			http.Error(w, "Items information not found", http.StatusNotFound)
			log.Println("Items information not found for order:", orderID)
			return
		}

		fullOrder := struct {
			Order    *db.Order    `json:"order"`
			Delivery *db.Delivery `json:"delivery"`
			Payment  *db.Payment  `json:"payment"`
			Items    []db.Item    `json:"items"`
		}{
			Order:    order,
			Delivery: delivery,
			Payment:  payment,
			Items:    items,
		}

		log.Printf("Responding with order details: %+v\n", fullOrder)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(fullOrder); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			log.Println("Failed to encode response for order:", orderID)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
