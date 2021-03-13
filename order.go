package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Order struct {
	ID           int     `json:"id"`
	UserID       int     `json:"userId"`
	AddressID    int     `json:"addressId`
	SessionID    string  `json:"sessionId"`
	Completed    bool    `json:"completed"`
	Subtotal     float32 `json:"subtotal"`
	Tax          float32 `json:"tax"`
	DeliveryFee  float32 `json:"deliveryFee"`
	ServiceFee   float32 `json:"serviceFee"`
	DeliveryTime string  `json:"deliveryTime"`
}

type orderService struct {
	db *sql.DB
	us userService
}

func (os orderService) findByUser(ctx context.Context) ([]*Order, error) {
	uid, err := os.us.idFromSession(ctx)
	if err != nil {
		return nil, err
	}
	q := `
		SELECT
			order_id, user_id, completed,
			COALESCE(address_id, 0),
			COALESCE(session_id, ''),
			COALESCE(subtotal, 0),
			COALESCE(tax, 0),
			COALESCE(delivery_fee, 0),
			COALESCE(service_fee, 0),
			COALESCE(delivery_time, '')
		FROM orders
		WHERE user_id = ?`
	rows, err := os.db.Query(q, uid)
	if err != nil {
		return nil, fmt.Errorf("finding orders by user ID: %v", err)
	}
	defer rows.Close()
	var orders []*Order
	for rows.Next() {
		var o Order
		err := rows.
			Scan(&o.ID, &o.UserID, &o.Completed, &o.AddressID, &o.SessionID,
				&o.Subtotal, &o.Tax, &o.DeliveryFee, &o.ServiceFee,
				&o.DeliveryTime)
		if err != nil {
			return nil, fmt.Errorf("reading order found by user ID: %v", err)
		}
		orders = append(orders, &o)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("reading orders found by user ID: %v", err)
	}
	return orders, nil
}

func (os orderService) create(o *Order) error {
	q := "INSERT INTO orders (user_id) VALUES (?)"
	stmt, err := os.db.Prepare(q)
	if err != nil {
		return fmt.Errorf("preparing order insertion query: %v", err)
	}
	defer stmt.Close()
	res, err := stmt.Exec(o.UserID)
	if err != nil {
		return fmt.Errorf("executing order insertion query: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting order ID: %v", err)
	}
	o.ID = int(id)
	return nil
}

func (os orderService) currentID(ctx context.Context) (int, error) {
	var id int
	uid, err := os.us.idFromSession(ctx)
	if err != nil {
		return id, err
	}
	q := `
		SELECT order_id
		FROM orders
		WHERE user_id = ? AND completed = FALSE
		ORDER BY created_at DESC`
	err = os.db.QueryRow(q, uid).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			o := &Order{UserID: uid}
			e := os.create(o)
			if e != nil {
				return o.ID, fmt.Errorf("getting current order: %v", e)
			}
			return o.ID, nil
		}
		return id, fmt.Errorf("finding current order: %v", err)
	}
	return id, nil
}

func (os orderService) checkOut(id int) (*Order, error) {
	return nil, nil
}

func (os orderService) place(id int) (*Order, error) {
	return nil, nil
}
