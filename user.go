package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/alexedwards/scs/v2"
	"github.com/graphql-go/graphql"
)

type User struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
}

var UserType = graphql.NewObject(
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

type UserService struct {
	db *sql.DB
	sm *scs.SessionManager
}

func (us UserService) FindByPhone(phone string) (*User, error) {
	var u User
	err := us.db.QueryRow("SELECT * FROM user WHERE phone = ?", phone).
		Scan(&u.FirstName, &u.LastName, &u.Phone, &u.Email)
	if err != nil {
		return nil, fmt.Errorf("finding user by phone: %v", err)
	}
	return &u, nil
}

func (us UserService) signUp(u *User, code string, ctx context.Context) error {
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

func (us UserService) logIn(phone, code string, ctx context.Context) (*User, error) {
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

func (us UserService) logOut(ctx context.Context) error {
	return us.sm.Destroy(ctx)
}

func (us UserService) phoneFromSession(ctx context.Context) string {
	return us.sm.GetString(ctx, "phone")
}

func createSession(phone string, sm *scs.SessionManager, ctx context.Context) error {
	err := sm.RenewToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to renew session token: %v", err)
	}
	sm.Put(ctx, "phone", phone)
	return nil
}

func me(app *App) *graphql.Field {
	return &graphql.Field{
		Type: UserType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			phone := app.users.phoneFromSession(p.Context)
			if phone == "" {
				return nil, errors.New("no session found")
			}
			u, err := app.users.FindByPhone(phone)
			if err != nil {
				return nil, err
			}
			return u, nil
		},
	}
}

func signUp(app *App) *graphql.Field {
	return &graphql.Field{
		Type: UserType,
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
			err := app.users.signUp(u, p.Args["code"].(string), p.Context)
			if err != nil {
				return nil, err
			}
			return u, nil
		},
	}
}

func logIn(app *App) *graphql.Field {
	return &graphql.Field{
		Type: UserType,
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
			u, err := app.users.logIn(phone, code, p.Context)
			if err != nil {
				return nil, err
			}
			return u, nil
		},
	}
}

func logOut(app *App) *graphql.Field {
	return &graphql.Field{
		Type: graphql.Boolean,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			err := app.users.logOut(p.Context)
			if err != nil {
				return nil, fmt.Errorf("failed to destroy session: %v", err)
			}
			return true, nil
		},
	}
}
