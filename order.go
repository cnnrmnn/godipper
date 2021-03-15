package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/cnnrmnn/godipper/chilis"
	"github.com/graphql-go/graphql"
)

// An Order is an order of triple dippers.
type Order struct {
	ID            int             `json:"id"`
	UserID        int             `json:"userId"`
	SessionID     string          `json:"sessionId"`
	Address       *Address        `json:"addressId`
	TripleDippers []*TripleDipper `json:"tripleDippers"`
	Completed     bool            `json:"completed"`
	Subtotal      float32         `json:"subtotal"`
	Tax           float32         `json:"tax"`
	DeliveryFee   float32         `json:"deliveryFee"`
	ServiceFee    float32         `json:"serviceFee"`
	DeliveryTime  time.Time       `json:"deliveryTime"`
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
		WHERE completed = TRUE AND user_id = ?`
	rows, err := os.db.Query(q, uid)
	if err != nil {
		return nil, fmt.Errorf("finding orders by user ID: %v", err)
	}
	defer rows.Close()
	var orders []*Order
	for rows.Next() {
		o := Order{Address: &Address{}}
		err := rows.
			Scan(&o.ID, &o.UserID, &o.Completed, &o.Address.ID, &o.SessionID,
				&o.Subtotal, &o.Tax, &o.DeliveryFee, &o.ServiceFee,
				&o.DeliveryTime)
		if err != nil {
			return nil, fmt.Errorf("reading order found by user ID: %v", err)
		}
		o.Address, err = os.as.findByID(o.Address.ID)
		if err != nil {
			return nil, fmt.Errorf("getting order address: %v", err)
		}
		o.TripleDippers, err = os.tds.findByOrder(o.ID)
		if err != nil {
			return nil, fmt.Errorf("getting order triple dippers: %v", err)
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
	o := Order{Address: &Address{}}
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
		WHERE completed = FALSE AND user_id = ?
		ORDER BY created_at DESC`
	err = os.db.QueryRow(q, uid).
		Scan(&o.ID, &o.UserID, &o.Completed, &o.Address.ID, &o.SessionID,
			&o.Subtotal, &o.Tax, &o.DeliveryFee, &o.ServiceFee,
			&o.DeliveryTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			no := &Order{UserID: uid}
			e := os.create(no)
			if e != nil {
				return nil, fmt.Errorf("getting current order: %v", e)
			}
			return no, nil
		}
		return nil, fmt.Errorf("finding current order: %v", err)
	}
	o.Address, err = os.as.findByID(o.Address.ID)
	if err != nil {
		return nil, fmt.Errorf("getting order address: %v", err)
	}
	o.TripleDippers, err = os.tds.findByOrder(o.ID)
	if err != nil {
		return nil, fmt.Errorf("getting order triple dippers: %v", err)
	}
	return &o, nil
}

// updateOrder updates the mutable fields in the database row corresponsing to
// the given order.
func (os orderService) updateOrder(o *Order) error {
	q := `
		UPDATE orders
		SET
			address_id = ?,
			session_id = ?,
			subtotal = ?,
			tax = ?,
			delivery_fee = ?,
			service_fee = ?,
			delivery_time = ?,
			completed = ?
		WHERE order_id = ?`
	stmt, err := os.db.Prepare(q)
	if err != nil {
		return fmt.Errorf("preparing order update query: %v", err)
	}
	_, err = stmt.Exec(o.Address.ID, o.SessionID, o.Subtotal, o.Tax,
		o.DeliveryFee, o.ServiceFee, o.DeliveryTime, o.Completed, o.ID)
	if err != nil {
		return fmt.Errorf("executing order update query: %v", err)
	}
	return nil
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

// checkOut populates the current user's current order with information from
// Chilis and returns it.
func (os orderService) checkOut(ctx context.Context, aid int) (*Order, error) {
	o, err := os.current(ctx)
	if err != nil {
		return nil, err
	}
	tdrs, err := os.tds.findByOrder(o.ID)
	if err != nil {
		return nil, err
	}
	if len(tdrs) == 0 {
		return nil, errors.New("cart is empty")
	}

	sess, err := chilis.StartSession()
	if err != nil {
		return nil, err
	}
	a, err := os.as.findByID(aid)
	if err != nil {
		return nil, err
	}
	err = sess.SetLocation(a.Address)
	if err != nil {
		return nil, err
	}

	for _, td := range tdrs {
		err = sess.Cart(td)
		if err != nil {
			return nil, err
		}
	}

	u, err := os.us.me(ctx)
	if err != nil {
		return nil, err
	}
	info, err := sess.Checkout(u.Customer, a.Address)
	if err != nil {
		return nil, err
	}
	o.Subtotal = info.Subtotal
	o.Tax = info.Tax
	o.DeliveryFee = info.DeliveryFee
	o.ServiceFee = info.ServiceFee
	o.DeliveryTime = info.DeliveryTime
	o.Address.ID = aid
	o.SessionID = sess.ID
	err = os.updateOrder(o)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// place places and returns the current user's current order.
func (os orderService) place(ctx context.Context, pm *chilis.PaymentMethod) (*Order, error) {
	o, err := os.current(ctx)
	if err != nil {
		return nil, err
	}
	if o.SessionID == "" {
		return nil, errors.New("check out before placing an order")
	}
	sess, err := chilis.NewSession(o.SessionID)
	if err != nil {
		return nil, err
	}
	_, err = sess.Order(pm)
	if err != nil {
		return nil, err
	}

	o.Completed = true
	err = os.updateOrder(o)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// orderType is the GraphQL type for Order.
var orderType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Order",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"userId": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"sessionId": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"tripleDippers": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(tripleDipperType))),
			},
			"address": &graphql.Field{
				Type: graphql.NewNonNull(addressType),
			},
			"completed": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
			},
			"subtotal": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Float),
			},
			"tax": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Float),
			},
			"deliveryFee": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Float),
			},
			"serviceFee": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Float),
			},
			"deliveryTime": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
	},
)

// orders returns a GraphQL query field that resolves to the current user's
// completed orders.
func orders(svc *service) *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(orderType))),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return svc.order.findByUser(p.Context)
		},
	}
}

// currentOrder returns a GraphQL query field that resolves to the current user's
// current order.
func currentOrder(svc *service) *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewNonNull(orderType),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return svc.order.current(p.Context)
		},
	}
}

// checkOut returns a GraphQL mutation field that populates the current user's
// current order with information from Chili's given an address ID.
func checkOut(svc *service) *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewNonNull(orderType),
		Args: graphql.FieldConfigArgument{
			"addressId": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return svc.order.checkOut(p.Context, p.Args["addressId"].(int))
		},
	}
}

// placeOrder returns a GraphQL mutation field that places and resolves to the
// current user's current order.
func placeOrder(svc *service) *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewNonNull(orderType),
		Args: graphql.FieldConfigArgument{
			"number": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"cvv": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"name": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"month": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"year": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"zip": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			pm := &chilis.PaymentMethod{
				Number: p.Args["number"].(string),
				CVV:    p.Args["cvv"].(string),
				Name:   p.Args["name"].(string),
				Month:  p.Args["month"].(string),
				Year:   p.Args["year"].(string),
				Zip:    p.Args["zip"].(string),
			}
			return svc.order.place(p.Context, pm)
		},
	}
}
