package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/cnnrmnn/godipper/chilis"
)

// An Order is an order of triple dippers.
type Order struct {
	ID           int       `json:"id"`
	UserID       int       `json:"userId"`
	AddressID    int       `json:"addressId`
	SessionID    string    `json:"sessionId"`
	Completed    bool      `json:"completed"`
	Subtotal     float32   `json:"subtotal"`
	Tax          float32   `json:"tax"`
	DeliveryFee  float32   `json:"deliveryFee"`
	ServiceFee   float32   `json:"serviceFee"`
	DeliveryTime time.Time `json:"deliveryTime"`
}

// orderService implements the order interface. Its methods manage orders.
type orderService struct {
	db  *sql.DB
	as  addressService
	tds tripleDipperService
	us  userService
}

// findByUser retuirns a slice of orders associated with the current user.
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
			COALESCE(delivery_time,
				STR_TO_DATE('1970-01-01 00:00:01', '%Y-%m-%d %H:%i:%s'))
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

// create creates an order.
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

// current return the current user's current order. If the current user has no
// current order, it creates an order and returns it.
func (os orderService) current(ctx context.Context) (*Order, error) {
	var o *Order
	uid, err := os.us.idFromSession(ctx)
	if err != nil {
		return o, err
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
		WHERE completed = FALSE AND order_id = ?
		ORDER BY created_at DESC`
	err = os.db.QueryRow(q, uid).
		Scan(&o.ID, &o.UserID, &o.Completed, &o.AddressID, &o.SessionID,
			&o.Subtotal, &o.Tax, &o.DeliveryFee, &o.ServiceFee,
			&o.DeliveryTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			o = &Order{UserID: uid}
			e := os.create(o)
			if e != nil {
				return o, fmt.Errorf("getting current order: %v", e)
			}
			return o, nil
		}
		return o, fmt.Errorf("finding current order: %v", err)
	}
	return o, nil
}

// cart creates a triple dipper that belongs to the current user's current
// order.
func (os orderService) cart(td *TripleDipper, ctx context.Context) error {
	o, err := os.current(ctx)
	if err != nil {
		return err
	}
	td.OrderID = o.ID
	return os.tds.create(td)
}

func (os orderService) updateOrder(o *Order, info chilis.OrderInfo) error {
	q := `
		UPDATE orders
		SET
			session_id = ?,
			subtotal = ?,
			tax = ?,
			delivery_fee = ?,
			service_fee = ?,
			delivery_time = ?
		WHERE order_id = ?`
	stmt, err := os.db.Prepare(q)
	if err != nil {
		return fmt.Errorf("preparing order update query: %v", err)
	}
	_, err = stmt.Exec(o.SessionID, o.Subtotal, o.Tax, o.DeliveryFee,
		o.ServiceFee, o.DeliveryTime)
	if err != nil {
		return fmt.Errorf("executing order update query: %v", err)
	}
	return nil
}

// checkOut populates the current user's current order with information from
// Chilis and returns it.
func (os orderService) checkOut(ctx context.Context, aid int) (*Order, error) {
	o, err := os.current(ctx)
	if err != nil {
		return o, err
	}
	sess, err := chilis.StartSession()
	if err != nil {
		return o, err
	}
	a, err := os.as.findByID(aid)
	if err != nil {
		return o, err
	}
	err = sess.SetLocation(a.Address)
	if err != nil {
		return o, err
	}
	tdrs, err := os.tds.findByOrder(o.ID)
	if err != nil {
		return o, err
	}
	if len(tdrs) == 0 {
		return o, errors.New("cart is empty")
	}
	for _, td := range tdrs {
		err = sess.Cart(td)
		if err != nil {
			return o, err
		}
	}
	u, err := os.us.me(ctx)
	if err != nil {
		return o, err
	}
	info, err := sess.Checkout(u.Customer, a.Address)
	if err != nil {
		return o, err
	}
	err = os.updateOrder(o, info)
	if err != nil {
		return o, err
	}
	return o, nil
}

// place places and returns the current user's current order.
func (os orderService) place(ctx context.Context) (*Order, error) {
	return nil, nil
}
