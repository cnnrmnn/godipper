package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/graphql-go/graphql"
)

// An Address is a United States address that can receive food deliveries.
type Address struct {
	ID     int    `json:"id"`
	UserID int    `json:"userId"`
	Street string `json:"street"`
	Unit   string `json:"unit"`
	City   string `json:"city"`
	State  string `json:"state"`
	Zip    string `json:"zip"`
	Notes  string `json:"notes"`
}

// addressService implements the address interface. Its methods manage
// addresses.
type addressService struct {
	db *sql.DB
	us user
}

// findByID returns the address with the given ID or an error if no address has
// the given ID.
func (as addressService) findByID(id int) (*Address, error) {
	q := `
		SELECT address_id, user_id, street, unit, city, state, zip, notes
		FROM addresses
		WHERE address_id = ?`
	var a Address
	err := as.db.QueryRow(q, id).
		Scan(&a.ID, &a.UserID, &a.Street, &a.Unit, &a.City, &a.State, &a.Zip, &a.Notes)
	if err != nil {
		return nil, fmt.Errorf("finding address by ID: %v", err)
	}
	return &a, nil
}

// findByUser returns a slice of addresses associated with the current user.
func (as addressService) findByUser(ctx context.Context) ([]*Address, error) {
	uid, err := as.us.idFromSession(ctx)
	if err != nil {
		return nil, err
	}
	q := `
		SELECT address_id, user_id, street, unit, city, state, zip, notes
		FROM addresses
		WHERE user_id = ?`
	rows, err := as.db.Query(q, uid)
	if err != nil {
		return nil, fmt.Errorf("finding addresses by user ID: %v", err)
	}
	defer rows.Close()
	var addrs []*Address
	for rows.Next() {
		var a Address
		err := rows.Scan(&a.ID, &a.UserID, &a.Street, &a.Unit, &a.City, &a.State, &a.Zip, &a.Notes)
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

// create creates an address that belongs to the current user.
func (as addressService) create(a *Address, ctx context.Context) error {
	uid, err := as.us.idFromSession(ctx)
	if err != nil {
		return err
	}
	a.UserID = uid
	q := `
		INSERT INTO addresses (user_id, street, unit, city, state, zip, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	stmt, err := as.db.Prepare(q)
	if err != nil {
		return fmt.Errorf("failed to prepare address insertion query: %v", err)
	}
	defer stmt.Close()
	res, err := stmt.Exec(a.UserID, a.Street, a.Unit, a.City, a.State, a.Zip, a.Notes)
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

// destroy destroys an address given its ID provided that it belongs to the
// current user.
func (as addressService) destroy(id int, ctx context.Context) (*Address, error) {
	uid, err := as.us.idFromSession(ctx)
	if err != nil {
		return nil, err
	}
	a, err := as.findByID(id)
	if err != nil {
		return nil, errors.New("failed to find address")
	}
	if a.UserID != uid {
		return nil, errors.New("address doesn't belong to user")
	}
	q := `DELETE FROM addresses WHERE address_id = ?`
	_, err = as.db.Exec(q, id)
	if err != nil {
		return nil, fmt.Errorf("destroying address by ID: %v", err)
	}
	return a, nil
}

// addressType is the GraphQL type for Address.
var addressType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Address",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"userId": &graphql.Field{
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

// addresses returns a GraphQL query field that resolves to the list of
// addresses that belong to the current user.
func addresses(svc *service) *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewList(graphql.NewNonNull(addressType)),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			addrs, err := svc.address.findByUser(p.Context)
			if err != nil {
				return nil, err
			}
			return addrs, nil
		},
	}
}

// createAddress returns a GraphQL mutation field that creates an address that
// belongs to the current user and resolves to that address if successful.
func createAddress(svc *service) *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewNonNull(addressType),
		Args: graphql.FieldConfigArgument{
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
				Type: graphql.String,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			a := &Address{
				Street: p.Args["street"].(string),
				City:   p.Args["city"].(string),
				State:  p.Args["state"].(string),
				Zip:    p.Args["zip"].(string),
			}
			unit, ok := p.Args["unit"].(string)
			if ok {
				a.Unit = unit
			}
			notes, ok := p.Args["notes"].(string)
			if ok {
				a.Notes = notes
			}

			err := svc.address.create(a, p.Context)
			if err != nil {
				return nil, err
			}
			return a, nil
		},
	}
}

// destroyAddress returns a GraphQL mutation field that destroys the address
// with the given ID provided that it belongs to the current user and resolves
// to that address if successful.
func destroyAddress(svc *service) *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewNonNull(addressType),
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id := p.Args["id"].(int)
			a, err := svc.address.destroy(id, p.Context)
			if err != nil {
				return false, err
			}
			return a, nil
		},
	}
}
