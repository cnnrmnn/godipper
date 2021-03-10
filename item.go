package main

import (
	"database/sql"
	"fmt"
)

type Item struct {
	ID             int      `json:"id"`
	TripleDipperID int      `json:"tripleDipperId"`
	ValueID        int      `json:"valueId"`
	Value          string   `json:"value"`
	Extras         []*Extra `json:"extras"`
}

type itemService struct {
	db *sql.DB
	es extra
}

func (is itemService) findByTripleDipper(tdid int) ([]*Item, error) {
	q := `
		SELECT i.item_id, i.triple_dipper_id, i.item_value_id, iv.item_value
		FROM item i INNER JOIN item_value iv
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

func (is itemService) create(it *Item, tx *sql.Tx) error {
	q := "INSERT INTO item (triple_dipper_id, item_value_id) VALUES(?, ?)"
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
