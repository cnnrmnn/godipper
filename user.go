package main

import (
	"database/sql"
	"errors"
	"fmt"

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

func (us UserService) Create(u *User, code string) error {
	ok, err := checkToken(u.Phone, code)
	if err != nil {
		return errors.New("couldn't check verification code")
	}
	if !ok {
		return errors.New("verification code is invalid")
	}
	q := `INSERT INTO user (first_name, last_name, phone, email)
				VALUES (?, ?, ?, ?)`
	stmt, err := us.db.Prepare(q)
	if err != nil {
		return errors.New("failed to create user")
	}
	res, err := stmt.Exec(u.FirstName, u.LastName, u.Phone, u.Email)
	if err != nil {
		return errors.New("failed to create user")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return errors.New("failed to get user ID")
	}
	u.ID = int(id)
	return nil
}

func (us UserService) Authenticate(phone, code string) (*User, error) {
	ok, err := checkToken(phone, code)
	if err != nil {
		return nil, errors.New("couldn't check verification code")
	}
	if !ok {
		return nil, errors.New("verification code is invalid")
	}
	var u User
	q := "SELECT * FROM user WHERE phone = ?"
	err = us.db.QueryRow(q, phone).
		Scan(&u.ID, &u.FirstName, &u.LastName, &u.Phone, &u.Email)
	if err != nil {
		return nil, errors.New("failed to get user ID")
	}
	return &u, nil
}

func me(app *App) *graphql.Field {
	return &graphql.Field{
		Type: UserType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id := app.sm.GetInt(p.Context, "id")
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
			err := app.users.Create(u, p.Args["code"].(string))
			if err != nil {
				return nil, err
			}
			err = app.sm.RenewToken(p.Context)
			if err != nil {
				return nil, errors.New("failed to renew session token")
			}
			app.sm.Put(p.Context, "id", u.ID)
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
			u, err := app.users.Authenticate(phone, code)
			if err != nil {
				return nil, err
			}
			err = app.sm.RenewToken(p.Context)
			if err != nil {
				return nil, errors.New("failed to renew session token")
			}
			app.sm.Put(p.Context, "id", u.ID)
			return u, nil
		},
	}
}
