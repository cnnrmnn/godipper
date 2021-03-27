package main

import (
	"database/sql"
	"fmt"

	"github.com/graphql-go/graphql"
)

// An Item is one of the three choices in a triple dipper.
type Item struct {
	ID             int      `json:"id"`
	TripleDipperID int      `json:"tripleDipperId"`
	ValueID        int      `json:"valueId"`
	Value          string   `json:"value"`
	Description    string   `json:"description"`
	ImagePath      string   `json:"imagePath"`
	Extras         []*Extra `json:"extras"`
}

// String returns a string representation of the item.
func (it Item) String() string {
	return it.Value
}

// ExtraValues returns a slice of the item's extras' values.
func (it Item) ExtraValues() []string {
	var vals []string
	for _, e := range it.Extras {
		vals = append(vals, e.Value)
	}
	return vals
}

// itemService implements the item interface. Its methods manage items.
type itemService struct {
	db *sql.DB
	es extra
}

// values returns a slice of all available item values.
func (is itemService) values() ([]*Item, error) {
	q := "SELECT item_value_id, item_value, description, image_path FROM item_values"
	rows, err := is.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("finding item values: %v", err)
	}
	defer rows.Close()
	var its []*Item
	for rows.Next() {
		var it Item
		err = rows.Scan(&it.ValueID, &it.Value, &it.Description, &it.ImagePath)
		if err != nil {
			return nil, fmt.Errorf("scanning item value: %v", err)
		}
		exs, err := is.es.values(it.ValueID)
		if err != nil {
			return nil, fmt.Errorf("finding item extra values: %v", err)
		}
		it.Extras = exs
		its = append(its, &it)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("reading item values: %v", err)
	}
	return its, nil
}

// findByTripleDipper returns a slice of items that belong to the triple dipper
// with the given ID.
func (is itemService) findByTripleDipper(tdid int) ([]*Item, error) {
	q := `
		SELECT i.item_id, i.triple_dipper_id, i.item_value_id, iv.item_value
		FROM items i INNER JOIN item_values iv
		ON i.item_value_id = iv.item_value_id
		WHERE i.triple_dipper_id = ?`
	rows, err := is.db.Query(q, tdid)
	if err != nil {
		return nil, fmt.Errorf("finding item by triple dipper ID: %v", err)
	}
	defer rows.Close()
	var its []*Item
	for rows.Next() {
		var it Item
		err = rows.Scan(&it.ID, &it.TripleDipperID, &it.ValueID, &it.Value)
		if err != nil {
			return nil, fmt.Errorf("reading item found by triple dipper ID: %v", err)
		}
		exts, err := is.es.findByItem(it.ID)
		if err != nil {
			return nil, fmt.Errorf("finding extras associated with item: %v", err)
		}
		it.Extras = exts
		its = append(its, &it)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("reading items found by triple dipper ID: %v", err)
	}
	return its, nil
}

// create creates the given item in the given transaction.
func (is itemService) create(it *Item, tx *sql.Tx) error {
	q := "INSERT INTO items (triple_dipper_id, item_value_id) VALUES(?, ?)"
	stmt, err := tx.Prepare(q)
	if err != nil {
		return fmt.Errorf("preparing item insertion query: %v", err)
	}
	res, err := stmt.Exec(it.TripleDipperID, it.ValueID)
	if err != nil {
		return fmt.Errorf("executing item insertion query: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting item ID: %v", err)
	}
	it.ID = int(id)
	for _, e := range it.Extras {
		e.ItemID = it.ID
		if err := is.es.create(e, tx); err != nil {
			return fmt.Errorf("inserting item extras: %v", err)
		}
	}
	return nil
}

// destroy destroys the items with the given triple dipper ID.
func (is itemService) destroy(tdid int, tx *sql.Tx) error {
	q := "SELECT item_id FROM items WHERE triple_dipper_id = ?"
	rows, err := is.db.Query(q, tdid)
	if err != nil {
		return fmt.Errorf("finding item IDs: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var iid int
		err := rows.Scan(&iid)
		if err != nil {
			return fmt.Errorf("reading item ID: %v", err)
		}
		err = is.es.destroy(iid, tx)
		if err != nil {
			return fmt.Errorf("destroying item extra: %v", err)
		}
	}
	q = "DELETE FROM items WHERE triple_dipper_id = ?"
	stmt, err := tx.Prepare(q)
	if err != nil {
		return fmt.Errorf("preparing item deletion query: %v", err)
	}
	_, err = stmt.Exec(tdid)
	if err != nil {
		return fmt.Errorf("executing item deletion query: %v", err)
	}
	return nil
}

// itemType is the GraphQL type for Item.
var itemType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Item",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"tripleDipperId": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"valueId": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"value": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"extras": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(extraType))),
			},
		},
	},
)

// itemValueType is the GraphQL type for an item value.
var itemValueType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "ItemValue",
		Fields: graphql.Fields{
			"valueId": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"value": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"description": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"imagePath": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"extras": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(extraValueType))),
			},
		},
	},
)

// itemInputType is the GraphQL input type for Item.
var itemInputType = graphql.NewInputObject(
	graphql.InputObjectConfig{
		Name: "ItemInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"valueId": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"extras": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.Int))),
			},
		},
	},
)

// itemValues returns a GraphQL query field that resolves to a list of
// available item values.
func itemValues(svc *service) *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(itemValueType))),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return svc.item.values()
		},
	}
}
