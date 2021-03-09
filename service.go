package main

import "context"

// service defines interface types for services used by GraphQL resolvers
// throughout the application.
type service struct {
	user interface {
		findByID(id int) (*User, error)
		findByPhone(phone string) (*User, error)
		signUp(u *User, code string, ctx context.Context) error
		logIn(phone, code string, ctx context.Context) (*User, error)
		logOut(ctx context.Context) error
		idFromSession(ctx context.Context) int
	}
}
