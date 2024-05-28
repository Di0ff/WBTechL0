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
	if cachedOrder, found := c.cache.Get(orderID); found {
		order := cachedOrder.(db.Order)
		log.Println("Order found in cache:", orderID)
		return &order, nil
	}

	order, err := c.fetchOrderFromDB(orderID)
	if err != nil {
		return nil, err
	}

	c.cache.Set(orderID, *order, cache.DefaultExpiration)
	log.Println("Order fetched from DB and added to cache:", orderID)
	return order, nil
}

func (c *Cache) GetDelivery(orderID string) (*db.Delivery, error) {
	if cachedDelivery, found := c.cache.Get(orderID + ":delivery"); found {
		delivery := cachedDelivery.(db.Delivery)
		return &delivery, nil
	}

	delivery, err := c.fetchDeliveryFromDB(orderID)
	if err != nil {
		return nil, err
	}

	c.cache.Set(orderID+":delivery", *delivery, cache.DefaultExpiration)
	return delivery, nil
}

func (c *Cache) GetPayment(orderID string) (*db.Payment, error) {
	if cachedPayment, found := c.cache.Get(orderID + ":payment"); found {
		payment := cachedPayment.(db.Payment)
		return &payment, nil
	}

	payment, err := c.fetchPaymentFromDB(orderID)
	if err != nil {
		return nil, err
	}

	c.cache.Set(orderID+":payment", *payment, cache.DefaultExpiration)
	return payment, nil
}

func (c *Cache) GetItems(orderID string) ([]db.Item, error) {
	if cachedItems, found := c.cache.Get(orderID + ":items"); found {
		items := cachedItems.([]db.Item)
		return items, nil
	}

	items, err := c.fetchItemsFromDB(orderID)
	if err != nil {
		return nil, err
	}

	c.cache.Set(orderID+":items", items, cache.DefaultExpiration)
	return items, nil
}

func (c *Cache) fetchOrderFromDB(orderID string) (*db.Order, error) {
	var order db.Order
	err := c.db.QueryRow(`
        SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        FROM orders
        WHERE order_uid = $1`, orderID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SMID, &order.DateCreated, &order.OOFShard)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (c *Cache) fetchDeliveryFromDB(orderID string) (*db.Delivery, error) {
	var delivery db.Delivery
	err := c.db.QueryRow(`SELECT order_uid, name, phone, zip, city, address, region, email FROM delivery WHERE order_uid = $1`, orderID).Scan(
		&delivery.OrderUID, &delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address, &delivery.Region, &delivery.Email)
	if err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (c *Cache) fetchPaymentFromDB(orderID string) (*db.Payment, error) {
	var payment db.Payment
	err := c.db.QueryRow(`SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee, order_uid FROM payment WHERE transaction = $1`, orderID).Scan(
		&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount, &payment.PaymentDt, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee, &payment.OrderUID)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (c *Cache) fetchItemsFromDB(orderID string) ([]db.Item, error) {
	rows, err := c.db.Query(`SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status, order_uid FROM items WHERE order_uid = $1`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []db.Item
	for rows.Next() {
		var item db.Item
		if err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NMID, &item.Brand, &item.Status, &item.OrderUID); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (c *Cache) LoadCacheFromDB() error {
	rows, err := c.db.Query(`SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var order db.Order
		err := rows.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature, &order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SMID, &order.DateCreated, &order.OOFShard)
		if err != nil {
			return err
		}
		c.cache.Set(order.OrderUID, order, cache.DefaultExpiration)
		log.Println("Order loaded into cache from DB:", order.OrderUID)
	}
	return nil
}
