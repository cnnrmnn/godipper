package main

import (
	"database/sql"
	"fmt"
)

type Extra struct {
	ID      int    `json:"id"`
	ItemID  int    `json:"itemId"`
	ValueID int    `json:"valueId"`
	Value   string `json:"value"`
}

type extraService struct {
	db *sql.DB
}

func (es extraService) findByItem(iid int) ([]*Extra, error) {
	q := `
		SELECT e.extra_id, e.item_id, e.extra_value_id, ev.extra_value
		FROM extras e INNER JOIN extra_values ev
		ON e.extra_value_id = ev.extra_value_id
		WHERE e.item_id = ?`
	rows, err := es.db.Query(q, iid)
	if err != nil {
		return nil, fmt.Errorf("finding extra by item ID: %v", err)
	}
	defer rows.Close()
	var exts []*Extra
	for rows.Next() {
		var e Extra
		err = rows.Scan(&e.ID, &e.ItemID, &e.ValueID, &e.Value)
		if err != nil {
			return nil, fmt.Errorf("reading extra found by item ID: %v", err)
		}
		exts = append(exts, &e)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("reading extras found by item ID: %v", err)
	}
	return exts, nil
}

func (es extraService) create(e *Extra, tx *sql.Tx) error {
	q := "INSERT INTO extras (item_id, extra_value_id) VALUES (?, ?)"
	stmt, err := tx.Prepare(q)
	if err != nil {
		return fmt.Errorf("preparing extra insertion query: %v", err)
	}
	res, err := stmt.Exec(e.ItemID, e.ValueID)
	if err != nil {
		return fmt.Errorf("executing extra insertion query: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting extra ID: %v", err)
	}
	e.ID = int(id)
	return nil
}
