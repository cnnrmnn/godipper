package main

import (
	"database/sql"
	"fmt"

	"github.com/graphql-go/graphql"
)

type Address struct {
	ID     int    `json:"id"`
	Street string `json:"street"`
	Unit   string `json:"unit"`
	City   string `json:"city"`
	State  string `json:"state"`
	Zip    string `json:"zip"`
	Notes  string `json:"notes"`
}

type addressService struct {
	db *sql.DB
}

func (as addressService) findByUser(uid int) ([]*Address, error) {
	q := `
		SELECT address_id, street, unit, city, state, zip
		FROM address
		WHERE user_id = ?`
	rows, err := as.db.Query(q, uid)
	if err != nil {
		return nil, fmt.Errorf("finding addresses by user ID: %v", err)
	}
	defer rows.Close()
	var addrs []*Address
	for rows.Next() {
		var a Address
		err := rows.Scan(&a.ID, &a.Street, &a.Unit, &a.City, &a.State, &a.Zip, &a.Notes)
		if err != nil {
			return nil, fmt.Errorf("reading address found by user ID: %v", err)
		}
		addrs = append(addrs, &a)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("reading addresses found by user ID: %v", err)
	}
	return addrs, nil
}

func (as addressService) create(a *Address, uid int) error {
	q := `
		INSERT INTO address (user_id, street, unit, city, state, zip)
		VALUES (?, ?, ?, ?, ?, ?)`
	stmt, err := as.db.Prepare(q)
	if err != nil {
		return fmt.Errorf("failed to prepare address insertion query: %v", err)
	}
	defer stmt.Close()
	res, err := stmt.Exec(uid, a.Street, a.Unit, a.City, a.State, a.Zip)
	if err != nil {
		return fmt.Errorf("failed to execute address insertion query: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get address ID: %v", err)
	}
	a.ID = int(id)
	return nil
}

func (as addressService) destroy(id int) error {
	q := `DELETE FROM address WHERE address_id = ?`
	_, err := as.db.Exec(q, id)
	if err != nil {
		return fmt.Errorf("destroying address by ID: %v", err)
	}
	return nil
}

var addressType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Address",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"street": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"unit": &graphql.Field{
				Type: graphql.String,
			},
			"city": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"state": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"zip": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"notes": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
	},
)

func addressByUser(svc *service) *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewList(graphql.NewNonNull(addressType)),
		Args: graphql.FieldConfigArgument{
			"userId": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			addrs, err := svc.address.findByUser(p.Args["userId"].(int))
			if err != nil {
				return nil, err
			}
			return addrs, nil
		},
	}
}

func createAddress(svc *service) *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewList(graphql.NewNonNull(addressType)),
		Args: graphql.FieldConfigArgument{
			"userId": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"street": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"unit": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"city": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"state": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"zip": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"notes": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			a := &Address{
				Street: p.Args["street"].(string),
				Unit:   p.Args["unit"].(string),
				City:   p.Args["city"].(string),
				State:  p.Args["state"].(string),
				Zip:    p.Args["zip"].(string),
				Notes:  p.Args["notes"].(string),
			}
			err := svc.address.create(a, p.Args["userId"].(int))
			if err != nil {
				return nil, err
			}
			return a, nil
		},
	}
}

func destroyAddress(svc *service) *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewNonNull(graphql.Boolean),
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			err := svc.address.destroy(p.Args["id"].(int))
			if err != nil {
				return false, err
			}
			return true, nil
		},
	}
}
