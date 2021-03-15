package main

import (
	"context"
	"database/sql"

	"github.com/cnnrmnn/godipper/chilis"
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
}

// extra defines the methods that should be implemented by the extra service.
type extra interface {
	values(ivid int) ([]*Extra, error)
	findByItem(iid int) ([]*Extra, error)
	create(e *Extra, tx *sql.Tx) error
	destroy(iid int, tx *sql.Tx) error
}

// item defines the methods that should be implemented by the item service.
type item interface {
	values() ([]*Item, error)
	findByTripleDipper(tdid int) ([]*Item, error)
	create(it *Item, tx *sql.Tx) error
	destroy(tdid int, tx *sql.Tx) error
}

// tripleDipper defines the methods that should be implemented by the
// tripleDipper service.
type tripleDipper interface {
	populate(td *TripleDipper) error
	findByID(id int) (*TripleDipper, error)
	findByOrder(oid int) ([]*TripleDipper, error)
	create(td *TripleDipper) error
	destroy(tdid int) error
}

// order defines the methods that should be implemented by the order service.
type order interface {
	populate(o *Order) error
	findByUser(ctx context.Context) ([]*Order, error)
	current(ctx context.Context) (*Order, error)
	create(o *Order) error
	cart(td *TripleDipper, ctx context.Context) error
	updateOrder(o *Order) error
	checkOut(ctx context.Context, aid int) (*Order, error)
	place(ctx context.Context, pm *chilis.PaymentMethod) (*Order, error)
}

// service defines interface types for services used by GraphQL resolvers
// throughout the application.
type service struct {
	user
	address
	extra
	item
	tripleDipper
	order
}
