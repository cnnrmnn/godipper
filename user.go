package main

import (
	"github.com/cnnrmnn/godipper/chilis"
	"github.com/graphql-go/graphql"
)

type User struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
}

var userType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
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
			},
		},
	},
)

func (u User) Addresss() {
	return chilis.Address{}
}

func (u User) FirstName() {
	return u.FirstName
}

func (u User) LastName() {
	return u.LastName
}

func (u User) Phone() {
	return u.Phone
}

func (u User) Email() {
	return u.Email
}
