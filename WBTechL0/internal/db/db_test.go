package db

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestAddOrder(t *testing.T) {
	dbConn, err := sql.Open("postgres", "postgres://admin:1111@localhost:5432/ordersWB?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	order := Order{
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
	delivery := Delivery{
		OrderUID: "testUID",
		Name:     "testName",
		Phone:    "+1234567890",
		Zip:      "123456",
		City:     "testCity",
		Address:  "testAddress",
		Region:   "testRegion",
		Email:    "test@test.com",
	}
	payment := Payment{
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
	items := []Item{
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

	err = AddOrder(dbConn, order, delivery, payment, items)
	if err != nil {
		t.Fatalf("Failed to add order: %v", err)
	}

	// Validate the inserted data
	var insertedOrder Order
	err = dbConn.QueryRow(`SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard 
                           FROM orders WHERE order_uid = $1`, order.OrderUID).Scan(
		&insertedOrder.OrderUID, &insertedOrder.TrackNumber, &insertedOrder.Entry, &insertedOrder.Locale,
		&insertedOrder.InternalSignature, &insertedOrder.CustomerID, &insertedOrder.DeliveryService,
		&insertedOrder.ShardKey, &insertedOrder.SMID, &insertedOrder.DateCreated, &insertedOrder.OOFShard)
	if err != nil {
		t.Fatalf("Failed to query inserted order: %v", err)
	}

	if insertedOrder.OrderUID != order.OrderUID {
		t.Errorf("Expected order UID %v, got %v", order.OrderUID, insertedOrder.OrderUID)
	}
}
