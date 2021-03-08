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
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
}

var UserType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
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

func (us UserService) FindByID(id int) (*User, error) {
	var u User
	err := us.db.QueryRow("SELECT * FROM user WHERE user_id = ?", id).
		Scan(&u.ID, &u.FirstName, &u.LastName, &u.Phone, &u.Email)
	if err != nil {
		return nil, fmt.Errorf("finding user by id: %v", err)
	}
	return &u, nil
}

func (us UserService) FindByPhone(phone string) (*User, error) {
	var u User
	err := us.db.QueryRow("SELECT * FROM user WHERE phone = ?", phone).
		Scan(&u.ID, &u.FirstName, &u.LastName, &u.Phone, &u.Email)
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
		Scan(&u.ID, &u.FirstName, &u.LastName, &u.Phone, &u.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated user: %v", err)
	}
	err = createSession(u.ID, us.sm, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}
	return &u, nil
}

func (us UserService) idFromSession(ctx context.Context) int {
	return us.sm.GetInt(ctx, "id")
}

func createSession(id int, sm *scs.SessionManager, ctx context.Context) error {
	err := sm.RenewToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to renew session token: %v", err)
	}
	sm.Put(ctx, "id", id)
	return nil
}

func me(app *App) *graphql.Field {
	return &graphql.Field{
		Type: UserType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id := app.users.idFromSession(p.Context)
			if id == 0 {
				return nil, errors.New("no session found")
			}
			u, err := app.users.FindByID(id)
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
