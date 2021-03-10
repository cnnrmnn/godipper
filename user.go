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
	ID        int    `json:"id"`
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

// findByID returns the user with the given ID or an an error if no user has
// the given ID
func (us userService) findByID(id int) (*User, error) {
	var u User
	q := `
		SELECT user_id, first_name, last_name, phone, email
		FROM user
		WHERE user_id = ?`
	err := us.db.QueryRow(q, id).
		Scan(&u.ID, &u.FirstName, &u.LastName, &u.Phone, &u.Email)
	if err != nil {
		return nil, fmt.Errorf("finding user by id: %v", err)
	}
	return &u, nil
}

// findByPhone returns the user with the given phone or an error if no user
// has the given phone.
func (us userService) findByPhone(phone string) (*User, error) {
	var u User
	q := `
		SELECT user_id, first_name, last_name, phone, email
		FROM user
		WHERE phone = ?`
	err := us.db.QueryRow(q, phone).
		Scan(&u.ID, &u.FirstName, &u.LastName, &u.Phone, &u.Email)
	if err != nil {
		return nil, fmt.Errorf("finding user by phone: %v", err)
	}
	return &u, nil
}

// me returns the user associated with the current session, if any, given the
// request context.
func (us userService) me(ctx context.Context) (*User, error) {
	id, err := us.idFromSession(ctx)
	if err != nil {
		return nil, err
	}
	u, err := us.findByID(id)
	if err != nil {
		return nil, err
	}
	return u, nil
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
	res, err := stmt.Exec(u.FirstName, u.LastName, u.Phone, u.Email)
	if err != nil {
		return fmt.Errorf("failed to execute user insertion query: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get user ID: %v", err)
	}
	u.ID = int(id)
	err = createSession(u.ID, us.sm, ctx)
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	return nil
}

// logIn creates a session for the user with the given phoneand returns that
// user provided that the verification code is valid.
func (us userService) logIn(phone, code string, ctx context.Context) (*User, error) {
	ok, err := checkToken(phone, code)
	if err != nil {
		return nil, fmt.Errorf("failed to check verification code: %v", err)
	}
	if !ok {
		return nil, errors.New("verification code is invalid")
	}
	u, err := us.findByPhone(phone)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated user: %v", err)
	}
	err = createSession(u.ID, us.sm, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}
	return u, nil
}

// logOut destroys the current session given the request context.
func (us userService) logOut(ctx context.Context) error {
	return us.sm.Destroy(ctx)
}

// idFromSession returns the ID associated with the current session given
// the request context.
func (us userService) idFromSession(ctx context.Context) (int, error) {
	id := us.sm.GetInt(ctx, "id")
	if id == 0 {
		return 0, errors.New("no session found")
	}
	return id, nil
}

// createSession creates a session for the given id.
func createSession(id int, sm *scs.SessionManager, ctx context.Context) error {
	err := sm.RenewToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to renew session token: %v", err)
	}
	sm.Put(ctx, "id", id)
	return nil
}

// userType is the GraphQL type for User.
var userType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"firstName": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"lastName": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"phone": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"email": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
	},
)

// me returns a GraphQL query field that resolves to the user associated with
// the current session.
func me(svc *service) *graphql.Field {
	return &graphql.Field{
		Type: userType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			u, err := svc.user.me(p.Context)
			if err != nil {
				return nil, err
			}
			return u, nil
		},
	}
}

// signUp returns a GraphQL mutation field that creates a user and resolves to
// that user if successful.
func signUp(svc *service) *graphql.Field {
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
			err := svc.user.signUp(u, p.Args["code"].(string), p.Context)
			if err != nil {
				return nil, err
			}
			return u, nil
		},
	}
}

// logIn returns a GraphQL mutation field that creates a new session and
// resolves to the user associated with that session if successful.
func logIn(svc *service) *graphql.Field {
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
			u, err := svc.user.logIn(phone, code, p.Context)
			if err != nil {
				return nil, err
			}
			return u, nil
		},
	}
}

// logout returns a GraphQL mutation field that destroys the currents session
// and resolves to a boolean value indicating if the operation was successful.
func logOut(svc *service) *graphql.Field {
	return &graphql.Field{
		Type: graphql.Boolean,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			err := svc.user.logOut(p.Context)
			if err != nil {
				return false, fmt.Errorf("failed to destroy session: %v", err)
			}
			return true, nil
		},
	}
}
