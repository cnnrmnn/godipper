package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/alexedwards/scs/v2"
	"github.com/graphql-go/graphql"
)

// A User is composed of all of the information associated with a user of the
// application.
type User struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
}

// userService implements the app.users interface. Its methods manage users and
// sessions.
type userService struct {
	db *sql.DB
	sm *scs.SessionManager
}

// findByPhone returns the user with the given phone or an error if no user
// has the given phone.
func (us userService) findByPhone(phone string) (*User, error) {
	var u User
	err := us.db.QueryRow("SELECT * FROM user WHERE phone = ?", phone).
		Scan(&u.FirstName, &u.LastName, &u.Phone, &u.Email)
	if err != nil {
		return nil, fmt.Errorf("finding user by Phone: %v", err)
	}
	return &u, nil
}

// signUp creates a row in the user table provided that the verification code
// valid. It creates a session for the user given the request context.
func (us userService) signUp(u *User, code string, ctx context.Context) error {
	ok, err := checkToken(u.Phone, code)
	if err != nil {
		return fmt.Errorf("couldn't check verification code: %v", err)
	}
	if !ok {
		return errors.New("verification code is invalid")
	}
	q := `INSERT INTO user (first_name, last_name, phone, email)
				VALUES (?, ?, ?, ?)`
	stmt, err := us.db.Prepare(q)
	if err != nil {
		return fmt.Errorf("failed to prepare user insertion query: %v", err)
	}
	_, err = stmt.Exec(u.FirstName, u.LastName, u.Phone, u.Email)
	if err != nil {
		return fmt.Errorf("failed to execute user insertion query: %v", err)
	}
	err = createSession(u.Phone, us.sm, ctx)
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	return nil
}

// logIn creates a session for the user with the given phone provided that the
// verification code is valid and returns that user.
func (us userService) logIn(phone, code string, ctx context.Context) (*User, error) {
	ok, err := checkToken(phone, code)
	if err != nil {
		return nil, fmt.Errorf("failed to check verification code: %v", err)
	}
	if !ok {
		return nil, errors.New("verification code is invalid")
	}
	var u User
	q := "SELECT * FROM user WHERE phone = ?"
	err = us.db.QueryRow(q, phone).
		Scan(&u.FirstName, &u.LastName, &u.Phone, &u.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated user: %v", err)
	}
	err = createSession(u.Phone, us.sm, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}
	return &u, nil
}

// logOut destroys the current session given the request context.
func (us userService) logOut(ctx context.Context) error {
	return us.sm.Destroy(ctx)
}

// phoneFromSession returns the phone associated with the current session given
// the request context.
func (us userService) phoneFromSession(ctx context.Context) string {
	return us.sm.GetString(ctx, "phone")
}

// createSession creates a session for the given phone.
func createSession(phone string, sm *scs.SessionManager, ctx context.Context) error {
	err := sm.RenewToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to renew session token: %v", err)
	}
	sm.Put(ctx, "phone", phone)
	return nil
}

// userType is the GraphQL type for User.
var userType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"firstName": &graphql.Field{
				Type: graphql.String,
			},
			"lastName": &graphql.Field{
				Type: graphql.String,
			},
			"phone": &graphql.Field{
				Type: graphql.String,
			},
			"email": &graphql.Field{
				Type: graphql.String,
			}},
	},
)

// me returns a GraphQL query field that resolves to the user associated with
// the current session.
func me(a *app) *graphql.Field {
	return &graphql.Field{
		Type: userType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			phone := a.users.phoneFromSession(p.Context)
			if phone == "" {
				return nil, errors.New("no session found")
			}
			u, err := a.users.findByPhone(phone)
			if err != nil {
				return nil, err
			}
			return u, nil
		},
	}
}

// signUp returns a GraphQL mutation field that creates a user and resolves to
// that user if successful.
func signUp(a *app) *graphql.Field {
	return &graphql.Field{
		Type: userType,
		Args: graphql.FieldConfigArgument{
			"firstName": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"lastName": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"phone": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"email": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"code": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			u := &User{
				FirstName: p.Args["firstName"].(string),
				LastName:  p.Args["lastName"].(string),
				Phone:     p.Args["phone"].(string),
				Email:     p.Args["email"].(string),
			}
			err := a.users.signUp(u, p.Args["code"].(string), p.Context)
			if err != nil {
				return nil, err
			}
			return u, nil
		},
	}
}

// logIn returns a GraphQL mutation field that creates a new session and
// resolves to the user associated with that session if successful.
func logIn(a *app) *graphql.Field {
	return &graphql.Field{
		Type: userType,
		Args: graphql.FieldConfigArgument{
			"phone": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"code": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			phone := p.Args["phone"].(string)
			code := p.Args["code"].(string)
			u, err := a.users.logIn(phone, code, p.Context)
			if err != nil {
				return nil, err
			}
			return u, nil
		},
	}
}

// logout returns a GraphQL mutation field that destroys the currents session
// and resolves to a boolean value indicating if the operation was successful.
func logOut(a *app) *graphql.Field {
	return &graphql.Field{
		Type: graphql.Boolean,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			err := a.users.logOut(p.Context)
			if err != nil {
				return false, fmt.Errorf("failed to destroy session: %v", err)
			}
			return true, nil
		},
	}
}
