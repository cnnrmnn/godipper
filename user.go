package main

import (
	"database/sql"
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

func user(app *App) *graphql.Field {
	return &graphql.Field{
		Type: UserType,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id, ok := p.Args["id"].(int)
			if ok {
				return app.users.ByID(id)
			}
			return nil, nil
		},
	}
}

func createUser(app *App) *graphql.Field {
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
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			u := &User{
				FirstName: p.Args["firstName"].(string),
				LastName:  p.Args["lastName"].(string),
				Phone:     p.Args["phone"].(string),
				Email:     p.Args["email"].(string),
			}
			return app.users.Create(u)
		},
	}

}

type UserService struct {
	db *sql.DB
}

func (us UserService) ByID(id int) (*User, error) {
	var u User
	err := us.db.QueryRow("SELECT * FROM user WHERE user_id = ?", id).
		Scan(&u.ID, &u.FirstName, &u.LastName, &u.Phone, &u.Email)
	if err != nil {
		return nil, fmt.Errorf("finding user by id: %v", err)
	}
	return &u, nil
}

func (us UserService) Create(u *User) (*User, error) {
	q := `INSERT INTO user (first_name, last_name, phone, email)
				VALUES (?, ?, ?, ?)`
	stmt, err := us.db.Prepare(q)
	if err != nil {
		return nil, nil
	}
	res, err := stmt.Exec(u.FirstName, u.LastName, u.Phone, u.Email)
	if err != nil {
		return nil, nil
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, nil
	}
	u.ID = int(id)
	return u, nil
}
