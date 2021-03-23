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
	Location      string          `json:"location"`
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

// populate populates the order's list of triple dippers and the order's
// address (if applicable).
func (ors orderService) populate(o *Order) error {
	var err error
	if o.Address.ID != 0 {
		o.Address, err = ors.as.findByID(o.Address.ID)
		if err != nil {
			return fmt.Errorf("getting order address: %v", err)
		}
	}
	o.TripleDippers, err = ors.tds.findByOrder(o.ID)
	if err != nil {
		return fmt.Errorf("getting order triple dippers: %v", err)
	}
	return nil
}

// findByUser retuirns a slice of orders associated with the current user.
func (ors orderService) findByUser(ctx context.Context) ([]*Order, error) {
	uid, err := ors.us.idFromSession(ctx)
	if err != nil {
		return nil, err
	}
	q := `
		SELECT
			order_id, user_id, completed,
			COALESCE(location, ''),
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
	rows, err := ors.db.Query(q, uid)
	if err != nil {
		return nil, fmt.Errorf("finding orders by user ID: %v", err)
	}
	defer rows.Close()
	var orders []*Order
	for rows.Next() {
		o := Order{Address: &Address{}}
		err := rows.
			Scan(&o.ID, &o.UserID, &o.Completed, &o.Location, &o.Address.ID,
				&o.SessionID, &o.Subtotal, &o.Tax, &o.DeliveryFee,
				&o.ServiceFee, &o.DeliveryTime)
		if err != nil {
			return nil, fmt.Errorf("reading order: %v", err)
		}
		err = ors.populate(&o)
		if err != nil {
			return nil, fmt.Errorf("reading order: %v", err)
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
func (ors orderService) create(o *Order) error {
	q := "INSERT INTO orders (user_id) VALUES (?)"
	stmt, err := ors.db.Prepare(q)
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
func (ors orderService) current(ctx context.Context) (*Order, error) {
	o := Order{Address: &Address{}}
	uid, err := ors.us.idFromSession(ctx)
	if err != nil {
		return nil, err
	}
	q := `
		SELECT
			order_id, user_id, completed,
			COALESCE(location, ''),
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
	err = ors.db.QueryRow(q, uid).
		Scan(&o.ID, &o.UserID, &o.Completed, &o.Location, &o.Address.ID,
			&o.SessionID, &o.Subtotal, &o.Tax, &o.DeliveryFee,
			&o.ServiceFee, &o.DeliveryTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			no := &Order{UserID: uid}
			e := ors.create(no)
			if e != nil {
				return nil, fmt.Errorf("getting current order: %v", e)
			}
			return no, nil
		}
		return nil, fmt.Errorf("finding current order: %v", err)
	}
        // This could be expensive for larger orders. Make it possible to turn
        // off order population when only the ID is needed.
	err = ors.populate(&o)
	if err != nil {
		return nil, fmt.Errorf("finding current order: %v", err)
	}
	return &o, nil
}

// updateOrder updates the mutable fields in the database row corresponsing to
// the given order.
func (ors orderService) updateOrder(o *Order) error {
	q := `
		UPDATE orders
		SET
			address_id = ?,
			location = ?,
			session_id = ?,
			subtotal = ?,
			tax = ?,
			delivery_fee = ?,
			service_fee = ?,
			delivery_time = ?,
			completed = ?
		WHERE order_id = ?`
	stmt, err := ors.db.Prepare(q)
	if err != nil {
		return fmt.Errorf("preparing order update query: %v", err)
	}
	_, err = stmt.Exec(o.Address.ID, o.Location, o.SessionID, o.Subtotal, o.Tax,
		o.DeliveryFee, o.ServiceFee, o.DeliveryTime, o.Completed, o.ID)
	if err != nil {
		return fmt.Errorf("executing order update query: %v", err)
	}
	err = ors.populate(o)
	if err != nil {
		return fmt.Errorf("updating order: %v", err)
	}
	return nil
}

// cart creates a triple dipper that belongs to the current user's current
// order.
func (ors orderService) cart(td *TripleDipper, ctx context.Context) error {
        o, err := ors.current(ctx)
	if err != nil {
		return err
	}
	td.OrderID = o.ID
	return ors.tds.create(td)
}

// uncart creates a triple dipper that belongs to the current user's current
// order.
func (ors orderService) uncart(tdid int, ctx context.Context) error {
	o, err := ors.current(ctx)
	if err != nil {
		return err
	}
	return ors.tds.destroy(tdid, o.ID)
}

// checkOut populates the current user's current order with information from
// Chilis and returns it.
func (ors orderService) checkOut(ctx context.Context, aid int) (*Order, error) {
	o, err := ors.current(ctx)
	if err != nil {
		return nil, err
	}
	tdrs, err := ors.tds.findByOrder(o.ID)
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
	a, err := ors.as.findByID(aid)
	if err != nil {
		return nil, err
	}
	if o.UserID != a.UserID {
		return nil, errors.New("address does not belong to current user")
	}
	err = sess.SetLocation(a.Address)
	if err != nil {
		return nil, err
	}

	// Tried to do this concurrently but Chili's server couldn't handle
	// concurrent requests.
	for _, td := range tdrs {
		err = sess.Cart(td)
		if err != nil {
			return nil, err
		}
	}

	u, err := ors.us.me(ctx)
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
	err = ors.updateOrder(o)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// place places and returns the current user's current order.
func (ors orderService) place(ctx context.Context, pm *chilis.PaymentMethod) (*Order, error) {
	o, err := ors.current(ctx)
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
	loc, err := sess.Order(pm)
	if err != nil {
		return nil, err
	}

	o.Location = loc
	o.Completed = true
	err = ors.updateOrder(o)
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
			"location": &graphql.Field{
				Type: graphql.String,
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

// addToCart returns a GraphQL mutation field that adds the given triple dipper
// to the current user's current order and resolves to that triple dipper.
func addToCart(svc *service) *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewNonNull(tripleDipperType),
		Args: graphql.FieldConfigArgument{
			"items": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(itemInputType))),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			var items []*Item
			for _, item := range p.Args["items"].([]interface{}) {
				iin := item.(map[string]interface{})
				ivid := iin["valueId"].(int)
				var extras []*Extra
				for _, ein := range iin["extras"].([]interface{}) {
					evid := ein.(int)
					extras = append(extras, &Extra{ValueID: evid})
				}
				items = append(items, &Item{ValueID: ivid, Extras: extras})
			}

			td := &TripleDipper{
				Items: items,
			}
			err := svc.order.cart(td, p.Context)
			if err != nil {
				return nil, err
			}
			return td, nil
		},
	}
}

// removeFromCart returns a GraphQL mutation field that removes the given
// triple dipper from the current user's current order and resolves to a
// boolean value reflecting the outcome of the operation.
func removeFromCart(svc *service) *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewNonNull(graphql.Boolean),
		Args: graphql.FieldConfigArgument{
			"tripleDipperId": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			err := svc.order.uncart(p.Args["tripleDipperId"].(int), p.Context)
			if err != nil {
				return false, err
			}
			return true, nil
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
