package cache

import (
	"database/sql"
	"testing"
	"time"

	"WBTechL0/internal/db"
	_ "github.com/lib/pq"
)

func TestCache(t *testing.T) {
	dbConn, err := sql.Open("postgres", "postgres://admin:1111@localhost:5432/ordersWB?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	orderCache := NewCache(dbConn)

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

	err = orderCache.LoadCacheFromDB()
	if err != nil {
		t.Fatalf("Failed to load cache from DB: %v", err)
	}

	cachedOrder, err := orderCache.GetOrder("testUID")
	if err != nil {
		t.Fatalf("Failed to get order from cache: %v", err)
	}

	if cachedOrder.OrderUID != order.OrderUID {
		t.Errorf("Expected order UID %v, got %v", order.OrderUID, cachedOrder.OrderUID)
	}
}
