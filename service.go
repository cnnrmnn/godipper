package main

import (
	"context"
	"database/sql"
)

// user defines the methods that should be implemented by the user service.
type user interface {
	findByID(id int) (*User, error)
	findByPhone(phone string) (*User, error)
	me(ctx context.Context) (*User, error)
	signUp(u *User, code string, ctx context.Context) error
	logIn(phone, code string, ctx context.Context) (*User, error)
	logOut(ctx context.Context) error
	idFromSession(ctx context.Context) (int, error)
}

// address defines the methods that should be implemented by the address
// service.
type address interface {
	findByID(id int) (*Address, error)
	findByUser(ctx context.Context) ([]*Address, error)
	create(a *Address, ctx context.Context) error
	destroy(id int, ctx context.Context) (*Address, error)
}

type extra interface {
	findByItem(iid int) ([]*Extra, error)
	create(e *Extra, tx *sql.Tx) error
}

type item interface {
	findByTripleDipper(tdid int) ([]*Item, error)
	create(it *Item, tx *sql.Tx) error
}

type tripleDipper interface {
	findByID(id int) (*TripleDipper, error)
	create(td *TripleDipper) error
}

// service defines interface types for services used by GraphQL resolvers
// throughout the application.
type service struct {
	user
	address
	extra
	item
	tripleDipper
}
