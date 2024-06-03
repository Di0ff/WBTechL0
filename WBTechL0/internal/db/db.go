package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type Order struct {
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	ShardKey          string    `json:"shardkey"`
	SMID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OOFShard          string    `json:"oof_shard"`
}

type Delivery struct {
	OrderUID string `json:"order_uid"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Zip      string `json:"zip"`
	City     string `json:"city"`
	Address  string `json:"address"`
	Region   string `json:"region"`
	Email    string `json:"email"`
}

type Payment struct {
	OrderUID     string `json:"order_uid"`
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int    `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	OrderUID    string `json:"order_uid"`
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NMID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

func AddOrder(db *sql.DB, order Order, delivery Delivery, payment Payment, items []Item) error {
	tx, err := db.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			log.Println("Error during transaction. Rolling back:", err)
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				log.Println("Error committing transaction:", err)
			} else {
				log.Println("Transaction committed successfully")

				// Проверка данных после коммита
				var insertedOrder Order
				err = db.QueryRow(`SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard 
                                   FROM orders WHERE order_uid = $1`, order.OrderUID).Scan(
					&insertedOrder.OrderUID, &insertedOrder.TrackNumber, &insertedOrder.Entry, &insertedOrder.Locale,
					&insertedOrder.InternalSignature, &insertedOrder.CustomerID, &insertedOrder.DeliveryService,
					&insertedOrder.ShardKey, &insertedOrder.SMID, &insertedOrder.DateCreated, &insertedOrder.OOFShard)
				if err != nil {
					log.Println("Error querying inserted order:", err)
				} else {
					log.Println("Inserted order:", insertedOrder)
				}

				// Проверка данных в таблице delivery после коммита
				var insertedDelivery Delivery
				err = db.QueryRow(`SELECT order_uid, name, phone, zip, city, address, region, email 
                                   FROM delivery WHERE order_uid = $1`, order.OrderUID).Scan(
					&insertedDelivery.OrderUID, &insertedDelivery.Name, &insertedDelivery.Phone, &insertedDelivery.Zip,
					&insertedDelivery.City, &insertedDelivery.Address, &insertedDelivery.Region, &insertedDelivery.Email)
				if err != nil {
					log.Println("Error querying inserted delivery:", err)
				} else {
					log.Println("Inserted delivery:", insertedDelivery)
				}
			}
		}
	}()

	// Отключение проверок внешних ключей
	_, err = tx.Exec(`SET session_replication_role = 'replica'`)
	if err != nil {
		log.Println("Error disabling foreign key checks:", err)
		return err
	}

	// Вставка данных в таблицу orders
	log.Printf("Inserting into orders table: %+v\n", order)
	result, err := tx.Exec(`
        INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerID, order.DeliveryService, order.ShardKey, order.SMID, order.DateCreated, order.OOFShard)
	if err != nil {
		log.Println("Error inserting into orders table:", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Println("Rows affected in orders table:", rowsAffected)

	// Вставка данных в таблицу delivery
	log.Printf("Inserting into delivery table: %+v\n", delivery)
	result, err = tx.Exec(`
        INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		delivery.OrderUID, delivery.Name, delivery.Phone, delivery.Zip, delivery.City, delivery.Address, delivery.Region, delivery.Email)
	if err != nil {
		log.Println("Error inserting into delivery table:", err)
		return err
	}

	rowsAffected, _ = result.RowsAffected()
	log.Println("Rows affected in delivery table:", rowsAffected)

	// Проверка вставленных данных сразу после вставки
	var insertedDelivery Delivery
	err = tx.QueryRow(`SELECT order_uid, name, phone, zip, city, address, region, email 
                       FROM delivery WHERE order_uid = $1`, delivery.OrderUID).Scan(
		&insertedDelivery.OrderUID, &insertedDelivery.Name, &insertedDelivery.Phone, &insertedDelivery.Zip,
		&insertedDelivery.City, &insertedDelivery.Address, &insertedDelivery.Region, &insertedDelivery.Email)
	if err != nil {
		log.Println("Error querying inserted delivery immediately after insertion:", err)
	} else {
		log.Println("Inserted delivery immediately after insertion:", insertedDelivery)
	}

	// Вставка данных в таблицу payment
	log.Printf("Inserting into payment table: %+v\n", payment)
	result, err = tx.Exec(`
        INSERT INTO payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee, order_uid)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		payment.Transaction, payment.RequestID, payment.Currency, payment.Provider, payment.Amount, payment.PaymentDt, payment.Bank, payment.DeliveryCost, payment.GoodsTotal, payment.CustomFee, order.OrderUID)
	if err != nil {
		log.Println("Error inserting into payment table:", err)
		return err
	}

	rowsAffected, _ = result.RowsAffected()
	log.Println("Rows affected in payment table:", rowsAffected)

	// Вставка данных в таблицу items
	for _, item := range items {
		log.Printf("Inserting into items table: %+v\n", item)
		result, err = tx.Exec(`
            INSERT INTO items (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status, order_uid)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name, item.Sale, item.Size, item.TotalPrice, item.NMID, item.Brand, item.Status, order.OrderUID)
		if err != nil {
			log.Println("Error inserting into items table:", err)
			return err
		}

		rowsAffected, _ = result.RowsAffected()
		log.Println("Rows affected in items table:", rowsAffected)
	}

	// Включение проверок внешних ключей
	_, err = tx.Exec(`SET session_replication_role = 'origin'`)
	if err != nil {
		log.Println("Error enabling foreign key checks:", err)
		return err
	}

	log.Println("Transaction committed successfully")
	return nil
}
