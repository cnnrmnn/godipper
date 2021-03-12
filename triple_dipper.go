package main

import (
	"database/sql"
	"fmt"
)

type TripleDipper struct {
	ID      int     `json:"id"`
	OrderID int     `json:"orderId"`
	Items   []*Item `json:"items"`
}

type tripleDipperService struct {
	db *sql.DB
	us user
	is item
}

func (tds tripleDipperService) findByID(id int) (*TripleDipper, error) {
	td := TripleDipper{ID: id}
	q := "SELECT order_id FROM triple_dipper where triple_dipper_id = ?"
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

func (tds tripleDipperService) create(td *TripleDipper) error {
	tx, err := tds.db.Begin()
	if err != nil {
		return fmt.Errorf("starting triple dipper transaction: %v", err)
	}
	q := "INSERT INTO triple_dipper (order_id) VALUES (?)"
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
