package cache

import (
	"WBTechL0/internal/db"
	"database/sql"
	"github.com/patrickmn/go-cache"
	"log"
	"time"
)

type Cache struct {
	cache *cache.Cache
	db    *sql.DB
}

func NewCache(db *sql.DB) *Cache {
	c := cache.New(5*time.Minute, 10*time.Minute)
	return &Cache{
		cache: c,
		db:    db,
	}
}

func (c *Cache) GetOrder(orderID string) (*db.Order, error) {
	log.Println("Fetching order from cache or DB:", orderID)
	if cachedOrder, found := c.cache.Get(orderID); found {
		order := cachedOrder.(db.Order)
		log.Println("Order found in cache:", orderID)
		return &order, nil
	}

	order, err := c.fetchOrderFromDB(orderID)
	if err != nil {
		log.Println("Error fetching order from DB:", err)
		return nil, err
	}

	c.cache.Set(orderID, *order, cache.DefaultExpiration)
	log.Println("Order fetched from DB and added to cache:", orderID)
	return order, nil
}

func (c *Cache) GetDelivery(orderID string) (*db.Delivery, error) {
	log.Println("Fetching delivery from cache or DB:", orderID)
	if cachedDelivery, found := c.cache.Get(orderID + ":delivery"); found {
		delivery := cachedDelivery.(db.Delivery)
		log.Println("Delivery found in cache:", orderID)
		return &delivery, nil
	}

	delivery, err := c.fetchDeliveryFromDB(orderID)
	if err != nil {
		log.Println("Error fetching delivery from DB:", err)
		return nil, err
	}

	c.cache.Set(orderID+":delivery", *delivery, cache.DefaultExpiration)
	log.Println("Delivery fetched from DB and added to cache:", orderID)
	return delivery, nil
}

func (c *Cache) GetPayment(orderID string) (*db.Payment, error) {
	log.Println("Fetching payment from cache or DB:", orderID)
	if cachedPayment, found := c.cache.Get(orderID + ":payment"); found {
		payment := cachedPayment.(db.Payment)
		log.Println("Payment found in cache:", orderID)
		return &payment, nil
	}

	payment, err := c.fetchPaymentFromDB(orderID)
	if err != nil {
		log.Println("Error fetching payment from DB:", err)
		return nil, err
	}

	c.cache.Set(orderID+":payment", *payment, cache.DefaultExpiration)
	log.Println("Payment fetched from DB and added to cache:", orderID)
	return payment, nil
}

func (c *Cache) GetItems(orderID string) ([]db.Item, error) {
	log.Println("Fetching items from cache or DB:", orderID)
	if cachedItems, found := c.cache.Get(orderID + ":items"); found {
		items := cachedItems.([]db.Item)
		log.Println("Items found in cache:", orderID)
		return items, nil
	}

	items, err := c.fetchItemsFromDB(orderID)
	if err != nil {
		log.Println("Error fetching items from DB:", err)
		return nil, err
	}

	c.cache.Set(orderID+":items", items, cache.DefaultExpiration)
	log.Println("Items fetched from DB and added to cache:", orderID)
	return items, nil
}

func (c *Cache) fetchOrderFromDB(orderID string) (*db.Order, error) {
	log.Println("Querying order from DB:", orderID)
	var order db.Order
	err := c.db.QueryRow(`
        SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        FROM orders
        WHERE order_uid = $1`, orderID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SMID, &order.DateCreated, &order.OOFShard)
	if err != nil {
		log.Println("Error querying order from DB:", err)
		return nil, err
	}
	return &order, nil
}

func (c *Cache) fetchDeliveryFromDB(orderID string) (*db.Delivery, error) {
	log.Println("Querying delivery from DB:", orderID)
	var delivery db.Delivery
	err := c.db.QueryRow(`SELECT order_uid, name, phone, zip, city, address, region, email FROM delivery WHERE order_uid = $1`, orderID).Scan(
		&delivery.OrderUID, &delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address, &delivery.Region, &delivery.Email)
	if err != nil {
		log.Println("Error querying delivery from DB:", err)
		return nil, err
	}
	log.Printf("Delivery fetched from DB: %+v\n", delivery)
	return &delivery, nil
}

func (c *Cache) fetchPaymentFromDB(orderID string) (*db.Payment, error) {
	log.Println("Querying payment from DB:", orderID)
	var payment db.Payment
	err := c.db.QueryRow(`SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee, order_uid FROM payment WHERE order_uid = $1`, orderID).Scan(
		&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount, &payment.PaymentDt, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee, &payment.OrderUID)
	if err != nil {
		log.Println("Error querying payment from DB:", err)
		return nil, err
	}
	return &payment, nil
}

func (c *Cache) fetchItemsFromDB(orderID string) ([]db.Item, error) {
	log.Println("Querying items from DB:", orderID)
	rows, err := c.db.Query(`SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status, order_uid FROM items WHERE order_uid = $1`, orderID)
	if err != nil {
		log.Println("Error querying items from DB:", err)
		return nil, err
	}
	defer rows.Close()

	var items []db.Item
	for rows.Next() {
		var item db.Item
		if err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NMID, &item.Brand, &item.Status, &item.OrderUID); err != nil {
			log.Println("Error scanning item row:", err)
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error iterating item rows:", err)
		return nil, err
	}
	return items, nil
}

func (c *Cache) LoadCacheFromDB() error {
	log.Println("Loading cache from DB")
	rows, err := c.db.Query(`SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders`)
	if err != nil {
		log.Println("Error querying orders for cache loading:", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var order db.Order
		err := rows.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature, &order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SMID, &order.DateCreated, &order.OOFShard)
		if err != nil {
			log.Println("Error scanning order row for cache:", err)
			return err
		}
		c.cache.Set(order.OrderUID, order, cache.DefaultExpiration)
		log.Println("Order loaded into cache from DB:", order.OrderUID)
	}
	return nil
}
