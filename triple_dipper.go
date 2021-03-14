package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/graphql-go/graphql"
)

// A TripleDipper is a Chili's Triple Dipper.
type TripleDipper struct {
	ID      int     `json:"id"`
	OrderID int     `json:"orderId"`
	Items   []*Item `json:"items"`
}

// ItemValues returns a slice of the triple dipper's items (that implement the
// chilis.Item interface).
func (td TripleDipper) ItemValues() []Item {
	var items []Item
	for _, it := range td.Items {
		items = append(items, *it)
	}
	return items
}

// tripleDipperService implements the tripleDipper interface. Its methods
// manage triple dippers.
type tripleDipperService struct {
	db *sql.DB
	us user
	is item
	os order
}

// findByID returns the triple dipper with the given ID or an error if no
// triple dipper has the given ID.
func (tds tripleDipperService) findByID(id int) (*TripleDipper, error) {
	td := TripleDipper{ID: id}
	q := "SELECT order_id FROM triple_dippers where triple_dipper_id = ?"
	err := tds.db.QueryRow(q, id).Scan(&td.OrderID)
	if err != nil {
		return nil, fmt.Errorf("finding triple dipper by ID: %v", err)
	}
	td.Items, err = tds.is.findByTripleDipper(id)
	if err != nil {
		return nil, fmt.Errorf("finding triple dipper items by ID: %v", err)
	}
	return &td, nil
}

// create creates a triple dipper.
func (tds tripleDipperService) create(td *TripleDipper) error {
	tx, err := tds.db.Begin()
	if err != nil {
		return fmt.Errorf("starting triple dipper transaction: %v", err)
	}
	q := "INSERT INTO triple_dippers (order_id) VALUES (?)"
	stmt, err := tx.Prepare(q)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("preparing triple dipper insertion query: %v", err)
	}
	res, err := stmt.Exec(td.OrderID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("executing triple dipper insertion query: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("getting triple dipper ID: %v", err)
	}
	td.ID = int(id)
	for _, it := range td.Items {
		it.TripleDipperID = td.ID
		if err := tds.is.create(it, tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("inserting triple dipper items: %v", err)
		}
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commiting triple dipper transaction: %v", err)
	}
	return nil
}

// cart creates a triple dipper that belongs to the current user's current
// order.
func (tds tripleDipperService) cart(td *TripleDipper, ctx context.Context) error {
	oid, err := tds.os.currentID(ctx)
	if err != nil {
		return err
	}
	td.OrderID = oid
	return tds.create(td)
}

// tripleDipperType is the GraphQL type for TripleDipper.
var tripleDipperType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "TripleDipper",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"orderId": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"items": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(itemType))),
			},
		},
	},
)

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
			err := svc.tripleDipper.cart(td, p.Context)
			if err != nil {
				return nil, err
			}
			// Populate value fields
			td, err = svc.tripleDipper.findByID(td.ID)
			if err != nil {
				return nil, err
			}
			return td, nil
		},
	}
}
