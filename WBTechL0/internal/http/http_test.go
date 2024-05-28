package http

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"WBTechL0/internal/cache"
	"WBTechL0/internal/db"
	_ "github.com/lib/pq"
)

func TestOrderHandler(t *testing.T) {
	dbConn, err := sql.Open("postgres", "postgres://admin:1111@localhost:5432/ordersWB?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	order := db.Order{
		OrderUID:          "testUID",
		TrackNumber:       "testTrack",
		Entry:             "testEntry",
		Locale:            "en",
		InternalSignature: "testSignature",
		CustomerID:        "testCustomer",
		DeliveryService:   "testService",
		ShardKey:          "testKey",
		SMID:              1,
		DateCreated:       time.Now(),
		OOFShard:          "1",
	}
	delivery := db.Delivery{
		OrderUID: "testUID",
		Name:     "testName",
		Phone:    "+1234567890",
		Zip:      "123456",
		City:     "testCity",
		Address:  "testAddress",
		Region:   "testRegion",
		Email:    "test@test.com",
	}
	payment := db.Payment{
		Transaction:  "testTransaction",
		RequestID:    "testRequest",
		Currency:     "USD",
		Provider:     "testProvider",
		Amount:       100,
		PaymentDt:    int(time.Now().Unix()),
		Bank:         "testBank",
		DeliveryCost: 50,
		GoodsTotal:   50,
		CustomFee:    0,
	}
	items := []db.Item{
		{
			ChrtID:      1,
			TrackNumber: "testTrack",
			Price:       50,
			RID:         "testRID",
			Name:        "testItem",
			Sale:        0,
			Size:        "M",
			TotalPrice:  50,
			NMID:        1,
			Brand:       "testBrand",
			Status:      1,
		},
	}

	err = db.AddOrder(dbConn, order, delivery, payment, items)
	if err != nil {
		t.Fatalf("Failed to add order: %v", err)
	}

	orderCache := cache.NewCache(dbConn)
	err = orderCache.LoadCacheFromDB()
	if err != nil {
		t.Fatalf("Failed to load cache from DB: %v", err)
	}

	req, err := http.NewRequest("GET", "/order/testUID", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orderID := r.URL.Path[len("/order/"):]
		order, err := orderCache.GetOrder(orderID)
		if err != nil {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}

		delivery, err := orderCache.GetDelivery(orderID)
		if err != nil {
			http.Error(w, "Delivery information not found", http.StatusNotFound)
			return
		}

		payment, err := orderCache.GetPayment(orderID)
		if err != nil {
			http.Error(w, "Payment information not found", http.StatusNotFound)
			return
		}

		items, err := orderCache.GetItems(orderID)
		if err != nil {
			http.Error(w, "Items information not found", http.StatusNotFound)
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

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(fullOrder); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	})
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, rr.Code)
	}
}
