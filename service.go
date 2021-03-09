package main

import "context"

type user interface {
	findByID(id int) (*User, error)
	findByPhone(phone string) (*User, error)
	signUp(u *User, code string, ctx context.Context) error
	logIn(phone, code string, ctx context.Context) (*User, error)
	logOut(ctx context.Context) error
	idFromSession(ctx context.Context) (int, error)
}

// service defines interface types for services used by GraphQL resolvers
// throughout the application.
type service struct {
	user
	address interface {
		findByID(id int) (*Address, error)
		findByUser(ctx context.Context) ([]*Address, error)
		create(a *Address, ctx context.Context) error
		destroy(id int, ctx context.Context) (*Address, error)
	}
}
