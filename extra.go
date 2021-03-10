package main

import (
	"database/sql"
	"fmt"
)

type Extra struct {
	ID     int  `json:"id"`
	ItemID int  `json:"itemId"`
	Value  byte `json:"value"`
}

type extraService struct {
	db *sql.DB
}

func (es extraService) findByItem(iid int) ([]*Extra, error) {
	q := "SELECT extra_id, extra FROM extra WHERE item_id = ?"
	rows, err := es.db.Query(q, iid)
	if err != nil {
		return nil, fmt.Errorf("finding extra by item ID: %v", err)
	}
	defer rows.Close()
	var exts []*Extra
	for rows.Next() {
		e := Extra{ItemID: iid}
		err = rows.Scan(&e.ID, &e.Value)
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
	q := "INSERT INTO extra (item_id, extra) VALUES (?, ?)"
	stmt, err := tx.Prepare(q)
	if err != nil {
		return fmt.Errorf("preparing extra insertion query: %v", err)
	}
	res, err := stmt.Exec(e.ItemID, e.Value)
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
