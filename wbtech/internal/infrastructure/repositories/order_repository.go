package repositories

import (
	"database/sql"
	"encoding/json"

	"github.com/agl/wbtech/internal/domain/entities"
	"github.com/agl/wbtech/pkg/logger"
)

type OrderRepository struct {
	db    *sql.DB
	cache map[string]*entities.Order
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	repo := &OrderRepository{
		db:    db,
		cache: make(map[string]*entities.Order),
	}
	repo.loadAllOrdersToCache()
	return repo
}

func (r *OrderRepository) loadAllOrdersToCache() {
	query := `SELECT order_uid FROM orders`
	rows, err := r.db.Query(query)
	if err != nil {
		logger.Log.Error("Failed to load orders for cache", "error", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			logger.Log.Error("Failed to scan order_uid for cache", "error", err)
			continue
		}
		order, err := r.GetOrderByID(orderUID)
		if err != nil || order == nil {
			logger.Log.Error("Failed to load order for cache", "order_uid", orderUID, "error", err)
			continue
		}
		r.cache[orderUID] = order
	}
	logger.Log.Info("Order cache initialized", "count", len(r.cache))
}

func (r *OrderRepository) GetOrderByID(orderUID string) (*entities.Order, error) {
	if order, ok := r.cache[orderUID]; ok {
		logger.Log.Info("Order found in cache", "order_uid", orderUID)
		return order, nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		logger.Log.Error("Failed to begin transaction", "error", err)
		return nil, err
	}

	var order entities.Order
	queryOrder := `SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders WHERE order_uid = $1`
	err = tx.QueryRow(queryOrder, orderUID).Scan(
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.ShardKey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Warn("Order not found", "order_uid", orderUID)

			return nil, nil
		}
		logger.Log.Error("Failed to select order", "order_uid", orderUID, "error", err)
		return nil, err
	}

	queryDelivery := `SELECT name, phone, zip, city, address, region, email FROM delivery WHERE order_uid = $1`
	err = tx.QueryRow(queryDelivery, orderUID).Scan(
		&order.Delivery.Name,
		&order.Delivery.Phone,
		&order.Delivery.Zip,
		&order.Delivery.City,
		&order.Delivery.Address,
		&order.Delivery.Region,
		&order.Delivery.Email,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Warn("Delivery not found", "order_uid", orderUID)
			return nil, nil
		}
		logger.Log.Error("Failed to select delivery", "order_uid", orderUID, "error", err)
		return nil, err
	}

	queryPayment := `SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payment WHERE order_uid = $1`
	err = tx.QueryRow(queryPayment, orderUID).Scan(
		&order.Payment.Transaction,
		&order.Payment.RequestID,
		&order.Payment.Currency,
		&order.Payment.Provider,
		&order.Payment.Amount,
		&order.Payment.PaymentDT,
		&order.Payment.Bank,
		&order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal,
		&order.Payment.CustomFee,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Warn("Payment not found", "order_uid", orderUID)

			return nil, nil
		}
		logger.Log.Error("Failed to select payment", "order_uid", orderUID, "error", err)
		return nil, err
	}

	queryItems := `SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid = $1`
	rows, err := tx.Query(queryItems, orderUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item entities.Item
		err := rows.Scan(
			&item.ChrtID,
			&item.TrackNumber,
			&item.Price,
			&item.Rid,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
		)
		if err != nil {
			logger.Log.Error("Failed to scan item", "order_uid", orderUID, "error", err)
			return nil, err
		}
		order.Items = append(order.Items, item)
	}
	if err = rows.Err(); err != nil {
		logger.Log.Error("Error iterating over items", "order_uid", orderUID, "error", err)
		return nil, err
	}

	logger.Log.Info("Order retrieved successfully", "order_uid", order.OrderUID)

	r.cache[orderUID] = &order

	return &order, nil
}

func (r *OrderRepository) StoreOrder(mshChan chan []byte) error {
	var order entities.Order
	for msg := range mshChan {
		if err := json.Unmarshal(msg, &order); err != nil {
			logger.Log.Error("Failed to unmarshal order", "error", err)
			return err
		}

		tx, err := r.db.Begin()
		if err != nil {
			logger.Log.Error("Failed to begin transaction", "error", err)
			return err
		}

		queryOrder := `INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
		_, err = tx.Exec(queryOrder,
			&order.OrderUID,
			&order.TrackNumber,
			&order.Entry,
			&order.Locale,
			&order.InternalSignature,
			&order.CustomerID,
			&order.DeliveryService,
			&order.ShardKey,
			&order.SmID,
			&order.DateCreated,
			&order.OofShard,
		)
		if err != nil {
			logger.Log.Error("Failed to insert order", "order_uid", order.OrderUID, "error", err)
			tx.Rollback()
			return err
		}

		queryDelivery := `INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
		_, err = tx.Exec(queryDelivery,
			&order.OrderUID,
			&order.Delivery.Name,
			&order.Delivery.Phone,
			&order.Delivery.Zip,
			&order.Delivery.City,
			&order.Delivery.Address,
			&order.Delivery.Region,
			&order.Delivery.Email,
		)
		if err != nil {
			logger.Log.Error("Failed to insert delivery", "order_uid", order.OrderUID, "error", err)
			tx.Rollback()
			return err
		}

		queryPayment := `INSERT INTO payment (order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
		_, err = tx.Exec(queryPayment,
			&order.OrderUID,
			&order.Payment.Transaction,
			&order.Payment.RequestID,
			&order.Payment.Currency,
			&order.Payment.Provider,
			&order.Payment.Amount,
			&order.Payment.PaymentDT,
			&order.Payment.Bank,
			&order.Payment.DeliveryCost,
			&order.Payment.GoodsTotal,
			&order.Payment.CustomFee,
		)
		if err != nil {
			logger.Log.Error("Failed to insert payment", "order_uid", order.OrderUID, "error", err)
			tx.Rollback()
			return err
		}

		for _, item := range order.Items {
			queryItem := `INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
			_, err = tx.Exec(queryItem,
				&order.OrderUID,
				&item.ChrtID,
				&item.TrackNumber,
				&item.Price,
				&item.Rid,
				&item.Name,
				&item.Sale,
				&item.Size,
				&item.TotalPrice,
				&item.NmID,
				&item.Brand,
				&item.Status,
			)
			if err != nil {
				logger.Log.Error("Failed to insert item", "order_uid", order.OrderUID, "chrt_id", item.ChrtID, "error", err)
				tx.Rollback()
				return err
			}
		}
		if err = tx.Commit(); err != nil {
			logger.Log.Error("Failed to commit transaction", "error", err)
			tx.Rollback()
			return err
		}

		r.cache[order.OrderUID] = &order
		logger.Log.Info("Order stored successfully", "order_uid", order.OrderUID)
	}

	return nil
}
